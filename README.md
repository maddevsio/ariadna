[![Go Report Card](http://goreportcard.com/badge/gen1us2k/ariadna)](http://goreportcard.com/report/gen1us2k/ariadna)
### Ariadna
Is the open-source geocoder built on top of ElasticSearch for fast geocoding and provides better search for CIS countries
### Work in progress. This may not work in your country. You are welcome for any issues, advices or other feedback

###What's a geocoder do anyway?

Geocoding is the process of transforming input text, such as an address, or a name of a place—to a location on the earth's surface.

![Ariadna](https://raw.githubusercontent.com/maddevsio/ariadna/master/img/geo.gif)


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
![Reverse](https://raw.githubusercontent.com/maddevsio/ariadna/master/img/reverse.gif)
### Prerequisites

* ElasticSearch
* PostgreSQL

Ariadna consists of 3 parts:
* Importer: OSM data importer to elastic search
* Updater: Download and import data
* WebUI for searching data
* Custom importer

### Install


```
  git clone git@github.com:gen1us2k/osm-geogoder.git
  cd osm-geogoder
  make depends
  make
```
### Configure

```
COMMANDS:
    import              Import OSM file to ElasticSearch
    update              Download OSM file and update index
    http                Run http server
    custom              Import custom data
    intersections       Process intersections only

GLOBAL OPTIONS:
   --config                                                                             Config file path
   --index_settings                                                                     ElasticSearch Index settings
   --custom_data                                                                        Custom data file path
   --es_index_name "addresses"                                                          Specify custom elasticsearch index name [$ARIADNA_ES_INDEX_NAME]
   --es_pg_conn_url "host=localhost user=geo password=geo dbname=geo sslmode=disable"   Specify custom PG connection URL [$ARIADNA_PG_CONN_URL]
   --es_url "http://localhost:9200"                                                     Custom url for elasticsearch e.g http://192.168.0.1:9200 [$ARIADNA_ES_HOST]
   --es_index_type "address"                                                            ElasticSearch index type [$ARIADNA_INDEX_TYPE]
   --filename "xxx"                                                                     filename for storing osm.pbf file [$ARIADNA_FILE_NAME]
   --download_url "xxx"                                                                 Geofabrik url to download file [$ARIADNA_DOWNLOAD_URL]
   --dont_import_intersections                                                          if checked, then ariadna won't import intersections [$ARIADNA_DONT_IMPORT_INTERSECTIONS]
   --help, -h                                                                           show help
   --version, -v                                                                        print the version
```
Elastic search index settings
```
cp index.json.example index.json
```
Change it for you
```
{
    "settings": {
        "analysis": {
            "filter": {
                "map_poi_filter": {
                    "type": "synonym",
                    "synonyms": [
                    	# all synonyms goes here
                    ]
                }
            },
            "analyzer": {
                "map_synonyms": {
                    "tokenizer": "standard",
                    "filter": [
                        "lowercase",
                        "map_poi_filter"
                    ]
                }
            }
        }
    },
    "mappings":
        {"address":
            {"properties":
                {
                    "centroid": {
                        "type": "geo_point" # need to reverse geocoding
                    }

                }
            }
        }
}
```

### Usage
First import data. Download it from geofabrik.de and run
```
$ ./ariadna import
```
Or you can specify download_url and file_name into settings and run
```
$ ./ariadna update
```
This creates elasticsearch index

### WebUI
```
$ ./ariadna http
```
Open http://localhost:8080 in your browser and enjoy

### Http API
There is http api for geocode and reverse geocode

1. /api/search/:query
2. /api/reverse/:lat/:lon

### Docker
To start Postgres, Elasticsearch and Ariadna run
```
$ cd ariadna-docker
$ cp ../index.json.example ./index.json
$ docker-compose up -d
$ docker-compose run --rm ariadna /go/bin/ariadna update
```
Open http://localhost:8080 in your browser and enjoy

### TODO
* Remove pg for searching intersections
* Test search for other countries
* Write some tests

## Roadmap

* Better code design
* Drop postgres dependency for searching crossroads
* Autocomplete
* More intelligent memory usage
* Add more sources for geocoding


### Contributing
1. Fork it ( https://github.com/maddevsio/ariadna/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request


### Pr, issues
You're welcome.

### NOTE
Tested only for my city and my country.


