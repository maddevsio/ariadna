package web

import "gopkg.in/olivere/elastic.v3"

var es *elastic.Client

func init() {
	es, _ = elastic.NewClient()
}
