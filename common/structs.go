package common

type IndexConfig struct {
	IndexVersion     int    `json:"index_version"`
	CurrentIndex     string `json:"current_index"`
	LastIndexVersion int    `json:"last_index_version"`
}

var (
	IC IndexConfig
	AC AppConfig
)

type AppConfig struct {
	IndexName               string
	PGConnString            string
	ElasticSearchHost       string
	IndexType               string
	FileName                string
	DownloadUrl             string
	DontImportIntersections bool
}
