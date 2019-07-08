package osm

import (
	"bytes"
	"fmt"
)

func (i *Importer) waysToElastic() error {
	buf, err := i.getWays()
	if err != nil {
		return err
	}
	return i.e.BulkWrite(buf)
}
func (i *Importer) getWays() (bytes.Buffer, error) {
	var buf bytes.Buffer
	for wayID, node := range i.handler.Ways {
		data, err := i.wayToJSON(node)
		if err != nil {
			return buf, err
		}
		meta := []byte(fmt.Sprintf(`{ "index": { "_id": "%d" } }%s`, wayID, "\n"))
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}
	return buf, nil
}
func (i *Importer) nodesToElastic() error {
	buf, err := i.getNodes()
	if err != nil {
		return err
	}
	return i.e.BulkWrite(buf)
}
func (i *Importer) getNodes() (bytes.Buffer, error) {
	var buf bytes.Buffer
	for nodeID, node := range i.handler.Nodes {
		data, err := i.nodeToJSON(node)
		if err != nil {
			return buf, err
		}
		meta := []byte(fmt.Sprintf(`{ "index": { "_id": "%d" } }%s`, nodeID, "\n"))
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}
	return buf, nil
}
