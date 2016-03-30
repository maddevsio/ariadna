package web

import (
	"gopkg.in/olivere/elastic.v3"
	"github.com/gen1us2k/ariadna/common"
)

var es *elastic.Client

func init() {
	es, _ = elastic.NewClient(
		elastic.SetURL(common.AC.ElasticSearchHost),
	)
}
