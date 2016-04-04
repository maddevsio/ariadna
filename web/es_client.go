package web

import (
	"github.com/gen1us2k/ariadna/common"
	"gopkg.in/olivere/elastic.v3"
)

var es *elastic.Client

func init() {
	es, _ = elastic.NewClient(
		elastic.SetURL(common.AC.ElasticSearchHost),
	)
}
