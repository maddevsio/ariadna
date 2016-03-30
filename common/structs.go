package common

type Config struct {
	IndexName               string `json:"index_name"`
	PGConnString            string `json:"pg_conn_string"`
	ElasticSearchHost	string `json:"elasticsearch_host"`
	IndexType               string `json:"index_type"`
	FileName                string `json:"file_name"`
	DownloadUrl             string `json:"download_url"`
	IndexVersion            int    `json:"index_version"`
	CurrentIndex            string `json:"current_index"`
	LastIndexVersion        int    `json:"last_index_version"`
	DontImportIntersections bool   `json:"dont_import_intersections"`
}

var C Config
