GeoGoder
===
More better geocoder for OpenStreetMap.
It able to search:
* Street + housenumber
* Road intersections
* Points of interest
* Microdictricts
* Search with auto replace from dictionary
* Reverse geocoding

Prerequisites
===
* ElasticSearch
* PostgreSQL

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
