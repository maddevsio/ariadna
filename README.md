[![Ariadna logo](https://user-images.githubusercontent.com/51479167/85124102-2cac4480-b24b-11ea-9e5d-b4441f8947b9.png)](https://blog.maddevs.io/ariadna-opensource-geocoder-832f149f2981)

[![Developed by Mad Devs](https://maddevs.io/badge-dark.svg)](https://maddevs.io/)
[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Go Report Card](http://goreportcard.com/badge/gen1us2k/ariadna)](http://goreportcard.com/report/gen1us2k/ariadna)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This is the open-source geocoder built on top of ElasticSearch for fast geocoding and providing better search for CIS countries.

You could find more information about Ariadna designing in our [blog](https://blog.maddevs.io/ariadna-opensource-geocoder-832f149f2981).

### What's a geocoding?

Geocoding is the process of transforming input text, such as an address, or a name of a place—to a location on the earth's surface.

### What can the Ariadne geocoder search for?

* Street + housenumber;
* Road intersections;
* Points of interest;
* Microdictricts;
* Addresses in microdistricts;
* Nearest villages and towns;
* Search with auto replace from dictionary;
* Reverse geocoding.


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

If you'd like to contribute, please fork the repository and make changes as you'd like. Pull requests are warmly welcome.

1. Fork it ( https://github.com/maddevsio/ariadna/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
