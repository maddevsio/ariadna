GeoGoder
===

Prerequisites
===
ElasticSearch
leveldb


Install'n'Run
===

```
  git clone git@github.com:gen1us2k/osm-geogoder.git
  cd osm-geogoder
  go get ./...
  go run importer.go kyrgyzstan-latest.osm.pbf
```

Search Data
===

```
  http://localhost:9200/addresses/_search=?q=QUERY
```
