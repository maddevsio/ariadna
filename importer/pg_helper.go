package importer

import (
	"database/sql"
	"fmt"
	"strconv"
	"github.com/gen1us2k/ariadna/common"
)

func RoadsToPg(Roads []JsonWay) {
	pg_db, err := sql.Open("postgres", common.C.PGConnString)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	defer pg_db.Close()

	_, err = pg_db.Query(`DROP TABLE IF EXISTS road;`)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	_, err = pg_db.Query(`DROP TABLE IF EXISTS road_intersection;`)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	_, err = pg_db.Query(`CREATE TABLE road (
			id serial not null primary key,
			node_id bigint not null,
			name varchar(255) null,
			old_name varchar(255) null,
			coords geometry
		);`)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	_, err = pg_db.Query(`create table road_intersection (
			id serial not null primary key,
			node_id bigint not null,
			name varchar(200) null,
			old_name varchar(255) null,
			coords geometry
		);`)

	if err != nil {
		Logger.Fatal(err.Error())
	}
	if err != nil {
		Logger.Fatal(err.Error())
	}

	Logger.Info("Creating tables")
	Logger.Info("Started populating tables with many roads")

	const insQuery = `INSERT INTO road (node_id, name, old_name, coords) values($1, $2, $3, ST_GeomFromText($4));`
	for _, road := range Roads {
		linestring := "LINESTRING("

		for _, point := range road.Nodes {
			linestring += fmt.Sprintf("%s %s,", strconv.FormatFloat(point.Lng(), 'f', 16, 64), strconv.FormatFloat(point.Lat(), 'f', 16, 64))
		}
		linestring = linestring[:len(linestring)-1]
		linestring += ")"
		insert_query, err := pg_db.Prepare(insQuery)

		if err != nil {
			panic(err)
		}
		defer insert_query.Close()

		name := ""
		if road.Tags["name"] != "" {
			name = road.Tags["name"]
		} else {
			name = road.Tags["addr:name"]
		}

		_, err = insert_query.Exec(road.ID, cleanAddress(name), cleanAddress(road.Tags["old_name"]), linestring)
		if err != nil {
			Logger.Fatal(err.Error())
		}
	}
	searchQuery := `
		INSERT INTO road_intersection( coords, name, old_name, node_id)
			(SELECT DISTINCT (ST_DUMP(ST_INTERSECTION(a.coords, b.coords))).geom AS ix,
			concat(a.name, ' ', b.name) as InterName,
			concat(a.old_name, ' ', b.old_name) as InterOldName,
			a.node_id + b.node_id
			FROM road a
			INNER JOIN road b
			ON ST_INTERSECTS(a.coords,b.coords)
			WHERE geometrytype(st_intersection(a.coords,b.coords)) = 'POINT'
		);
	`
	Logger.Info("Started searching intersections")
	_, err = pg_db.Query(searchQuery)

	if err != nil {
		Logger.Fatal(err.Error())
	}

}

func GetRoadIntersectionsFromPG() []JsonNode {
	var Nodes []JsonNode
	pg_db, err := sql.Open("postgres", common.C.PGConnString)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	defer pg_db.Close()
	rows, err := pg_db.Query("SELECT node_id, name, old_name, st_x((st_dump(coords)).geom) as lng, st_y((st_dump(coords)).geom) as lat from road_intersection")

	if err != nil {
		Logger.Fatal(err.Error())
	}
	for rows.Next() {
		var node PGNode
		rows.Scan(&node.ID, &node.Name, &node.OldName, &node.Lng, &node.Lat)
		tags := make(map[string]string)
		tags["name"] = node.Name
		tags["old_name"] = node.OldName
		jNode := JsonNode{node.ID, "node", node.Lat, node.Lng, tags}
		Nodes = append(Nodes, jNode)
	}
	return Nodes
}
