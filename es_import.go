package main

import (
	_ "github.com/lib/pq"
	"database/sql"
	"gopkg.in/olivere/elastic.v3"
	"fmt"
	"encoding/json"
)

type AddressData struct {
	City string			`json:"city"`
	Street string		`json:"street"`
	HouseNumber string	`json:"housenumber"`
	Centroid interface{} 	`json:"centroid"`
	Coords interface{}		`json:"coords"`
}
type Address struct {
	NodeID uint64		`json:"id"`
	AddressData *AddressData
}

func main() {
	// Set the Elasticsearch Host to Connect to
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}

	// Create an index
	_, err = client.CreateIndex("addresses").Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	pg_db, err := sql.Open("postgres", "host=localhost user=geo password=geo dbname=geo sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	rows, _ := pg_db.Query("select row_to_json(r.*) from (select node_id as id, city as city, street as street, housenumber as housenumber, st_asgeojson(centroid) as centroid, st_asgeojson(coords) as coords from addresses where city='Бишкек') r;")
	var dataString string
	var address Address
	var i uint64
	i = 0
	for rows.Next() {
		rows.Scan(&dataString)
		jerr := json.Unmarshal([]byte(dataString), &address)
		if jerr != nil {
			fmt.Println(jerr)
			fmt.Println(dataString)
			break
		}
		_, err := client.Index().
			Index("addresses").
			Type("address").
			Id(string(i)).
			BodyString(dataString).
			Do()

		if err != nil {
			fmt.Println(err)
			break
		}
		i += 1
	}

//	api.Domain = "localhost"
//	// api.Port = "9300"
//
//	indexer := core.NewBulkIndexerErrors(10, 60)
//	done := make(chan bool)
//	indexer.Run(done)
//
//	go func() {
//		for errBuf := range indexer.ErrorChannel {
//			// just blissfully print errors forever
//			fmt.Println(errBuf.Err)
//		}
//	}()
//	for i := 0; i < 20; i++ {
//		indexer.Index("twitter", "user", strconv.Itoa(i), "", nil, `{"name":"bob"}`, false)
//	}
//	done <- true
}

/*
select
row_to_json(r.*) from
    (select
        node_id as id,
        city as city,
        street as street,
        housenumber as housenumber,
        st_asgeojson(centroid) as centroid,
        st_asgeojson(coords) as coords
        from addresses)
    r;
 */