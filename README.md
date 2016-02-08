### GeoGoder
Its open-source geocoder built on top of ElasticSearch for fast geocoding and provide more better search for CIS countries

###What's a geocoder do anyway?

Geocoding is the process of transforming input text, such as an address, or a name of a place—to a location on the earth's surface.
// TODO: Add gif demo

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
* Make search for non Kyrgyzstan 
* Check search for non CIS countries
