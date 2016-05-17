package common

var (
	AC AppConfig
)

type AppConfig struct {
	IndexName               string
	PGConnString            string
	ElasticSearchHost       string
	IndexType               string
	FileName                string
	DownloadUrl             string
	ElasticSearchIndexUrl   string
	DontImportIntersections bool
}
