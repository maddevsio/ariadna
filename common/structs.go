package common

import "github.com/olivere/elastic"

var (
	AC AppConfig
	ES *elastic.Client
)

type (
	CustomData struct {
		ID   int64   `json:"id"`
		Name string  `json:"name"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
	}
	Custom []CustomData

	AppConfig struct {
		IndexName               string
		PGConnString            string
		ElasticSearchHost       string
		IndexType               string
		FileName                string
		DownloadUrl             string
		ElasticSearchIndexUrl   string
		DontImportIntersections bool
	}

	BadRequest struct {
		Error string `json:"error"`
	}
)
