[![Go Report Card](http://goreportcard.com/badge/gen1us2k/ariadna)](http://goreportcard.com/report/gen1us2k/ariadna)

### Ariadna

Is the open-source geocoder built on top of ElasticSearch for fast geocoding and provides better search for CIS countries

### What's a geocoding?

Geocoding is the process of transforming input text, such as an address, or a name of a place—to a location on the earth's surface.

It able to search:

* Street + housenumber
* Road intersections
* Points of interest
* Microdictricts
* Addresses in microdistricts
* Nearest villages and towns
* Search with auto replace from dictionary
* Reverse geocoding

###  What's reverse geocoding?

Reverse geocoding is the opposite, it transforms your current geographic location in to a list of places nearby.

### Prerequisites

* ElasticSearch

### Install 

```
go get -u github.com/maddevsio/ariadna
```

### Run

```
 go run main.go
 ```

### Configuration

You can use json or yaml files for configuration. Configuration example shown below. 

```
cat ariadna.yml
---                                                                                                                           elastic_index: addresses # index name for elasticsearch
elastic_urls:
  - http://localhost:9200   # array of elasticsearch addresses
osm_filename: kyrgyzstan-latest.osm.pbf # temporary filename for osm.pbf file downloaded from geofabrik        
osm_url: http://download.geofabrik.de/asia/kyrgyzstan-latest.osm.pbf  # Download url for osm.pdf file
index_settings: index.json   # Settings for index
import_country: Кыргызстан   # Country name to import
```

### Contributing
1. Fork it ( https://github.com/maddevsio/ariadna/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request


### Pr, issues
You're welcome.
