package osmpbf

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/qedus/osmpbf/OSMPBF"
)

// Decoder for Blob with OSMData (PrimitiveBlock)
type dataDecoder struct {
	q []interface{}
}

func (dec *dataDecoder) Decode(blob *OSMPBF.Blob) ([]interface{}, error) {
	dec.q = make([]interface{}, 0, 8000) // typical PrimitiveBlock contains 8k OSM entities

	data, err := getData(blob)
	if err != nil {
		return nil, err
	}

	primitiveBlock := &OSMPBF.PrimitiveBlock{}
	if err := proto.Unmarshal(data, primitiveBlock); err != nil {
		return nil, err
	}

	dec.parsePrimitiveBlock(primitiveBlock)
	return dec.q, nil
}

func (dec *dataDecoder) parsePrimitiveBlock(pb *OSMPBF.PrimitiveBlock) {
	for _, pg := range pb.GetPrimitivegroup() {
		dec.parsePrimitiveGroup(pb, pg)
	}
}

func (dec *dataDecoder) parsePrimitiveGroup(pb *OSMPBF.PrimitiveBlock, pg *OSMPBF.PrimitiveGroup) {
	dec.parseNodes(pb, pg.GetNodes())
	dec.parseDenseNodes(pb, pg.GetDense())
	dec.parseWays(pb, pg.GetWays())
	dec.parseRelations(pb, pg.GetRelations())
}

func (dec *dataDecoder) parseNodes(pb *OSMPBF.PrimitiveBlock, nodes []*OSMPBF.Node) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	dateGranularity := int64(pb.GetDateGranularity())

	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()

	for _, node := range nodes {
		id := node.GetId()
		lat := node.GetLat()
		lon := node.GetLon()

		latitude := 1e-9 * float64((latOffset + (granularity * lat)))
		longitude := 1e-9 * float64((lonOffset + (granularity * lon)))

		tags := extractTags(st, node.GetKeys(), node.GetVals())
		info := extractInfo(st, node.GetInfo(), dateGranularity)

		dec.q = append(dec.q, &Node{id, latitude, longitude, tags, info})
	}

}

func (dec *dataDecoder) parseDenseNodes(pb *OSMPBF.PrimitiveBlock, dn *OSMPBF.DenseNodes) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	dateGranularity := int64(pb.GetDateGranularity())
	ids := dn.GetId()
	lats := dn.GetLat()
	lons := dn.GetLon()
	di := dn.GetDenseinfo()

	tu := tagUnpacker{st, dn.GetKeysVals(), 0}
	var id, lat, lon int64
	var state denseInfoState
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		latitude := 1e-9 * float64((latOffset + (granularity * lat)))
		longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
		tags := tu.next()
		info := extractDenseInfo(st, &state, di, index, dateGranularity)

		dec.q = append(dec.q, &Node{id, latitude, longitude, tags, info})
	}
}

func (dec *dataDecoder) parseWays(pb *OSMPBF.PrimitiveBlock, ways []*OSMPBF.Way) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())

	for _, way := range ways {
		id := way.GetId()

		tags := extractTags(st, way.GetKeys(), way.GetVals())

		refs := way.GetRefs()
		var nodeID int64
		nodeIDs := make([]int64, len(refs))
		for index := range refs {
			nodeID = refs[index] + nodeID // delta encoding
			nodeIDs[index] = nodeID
		}

		info := extractInfo(st, way.GetInfo(), dateGranularity)

		dec.q = append(dec.q, &Way{id, tags, nodeIDs, info})
	}
}

// Make relation members from stringtable and three parallel arrays of IDs.
func extractMembers(stringTable []string, rel *OSMPBF.Relation) []Member {
	memIDs := rel.GetMemids()
	types := rel.GetTypes()
	roleIDs := rel.GetRolesSid()

	var memID int64
	members := make([]Member, len(memIDs))
	for index := range memIDs {
		memID = memIDs[index] + memID // delta encoding

		var memType MemberType
		switch types[index] {
		case OSMPBF.Relation_NODE:
			memType = NodeType
		case OSMPBF.Relation_WAY:
			memType = WayType
		case OSMPBF.Relation_RELATION:
			memType = RelationType
		}

		role := stringTable[roleIDs[index]]

		members[index] = Member{memID, memType, role}
	}

	return members
}

func (dec *dataDecoder) parseRelations(pb *OSMPBF.PrimitiveBlock, relations []*OSMPBF.Relation) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())

	for _, rel := range relations {
		id := rel.GetId()
		tags := extractTags(st, rel.GetKeys(), rel.GetVals())
		members := extractMembers(st, rel)
		info := extractInfo(st, rel.GetInfo(), dateGranularity)

		dec.q = append(dec.q, &Relation{id, tags, members, info})
	}
}

func extractInfo(stringTable []string, i *OSMPBF.Info, dateGranularity int64) Info {
	info := Info{Visible: true}

	if i != nil {
		info.Version = i.GetVersion()

		millisec := time.Duration(i.GetTimestamp()*dateGranularity) * time.Millisecond
		info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()

		info.Changeset = i.GetChangeset()

		info.Uid = i.GetUid()

		info.User = stringTable[i.GetUserSid()]

		if i.Visible != nil {
			info.Visible = i.GetVisible()
		}
	}

	return info
}

type denseInfoState struct {
	timestamp int64
	changeset int64
	uid       int32
	userSid   int32
}

func extractDenseInfo(stringTable []string, state *denseInfoState, di *OSMPBF.DenseInfo, index int, dateGranularity int64) Info {
	info := Info{Visible: true}

	versions := di.GetVersion()
	if len(versions) > 0 {
		info.Version = versions[index]
	}

	timestamps := di.GetTimestamp()
	if len(timestamps) > 0 {
		state.timestamp = timestamps[index] + state.timestamp
		millisec := time.Duration(state.timestamp*dateGranularity) * time.Millisecond
		info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()
	}

	changesets := di.GetChangeset()
	if len(changesets) > 0 {
		state.changeset = changesets[index] + state.changeset
		info.Changeset = state.changeset
	}

	uids := di.GetUid()
	if len(uids) > 0 {
		state.uid = uids[index] + state.uid
		info.Uid = state.uid
	}

	usersids := di.GetUserSid()
	if len(usersids) > 0 {
		state.userSid = usersids[index] + state.userSid
		info.User = stringTable[state.userSid]
	}

	visibleArray := di.GetVisible()
	if len(visibleArray) > 0 {
		info.Visible = visibleArray[index]
	}

	return info
}
