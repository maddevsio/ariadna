package main

import (
	"gopkg.in/olivere/elastic.v3"
	"fmt"
)

func main() {
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}
//	termQuery := elastic.NewTermQuery("user", "olivere")
	termQuery := elastic.NewTermQuery("street", "Киевская")
	searchResult, err := client.Search().
		Index("addresses").   // search in index "twitter"
		Query(termQuery).   // specify the query
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do()                // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Println(searchResult)
}