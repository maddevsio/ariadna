[![Go Report Card](http://goreportcard.com/badge/gen1us2k/ariadna)](http://goreportcard.com/report/gen1us2k/ariadna)
### Ariadna
Its open-source geocoder built on top of ElasticSearch for fast geocoding and provide more better search for CIS countries

###What's a geocoder do anyway?

Geocoding is the process of transforming input text, such as an address, or a name of a place—to a location on the earth's surface.

!(http://imgur.com/TPW4cOs)


It able to search:
* Street + housenumber
* Road intersections
* Points of interest
* Microdictricts
* Addresses in microdistricts
* Nearest villages and towns
* Search with auto replace from dictionary
* Reverse geocoding

### ... and a reverse geocoder, what's that?

Reverse geocoding is the opposite, it transforms your current geographic location in to a list of places nearby.

### Prerequisites

* ElasticSearch
* PostgreSQL

Ariadna consists of 3 parts:
* Importer: OSM data importer to elastic search
* Updater: Download and import data
* WebUI for searching data

### Install'n'Run


```
  git clone git@github.com:gen1us2k/osm-geogoder.git
  cd osm-geogoder
  make depends
  make 
  ./importer kyrgyzstan-latest.osm.pbf
```

### Search Data

Simple search can be accessed by
```
  http://localhost:9200/addresses/_search=?q=QUERY
```

### TODO
* Remove pg for searching intersections
* Test search for other countries
* Write some tests
