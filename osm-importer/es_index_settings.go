package importer

var ESSettings string = `{
    "settings": {
        "analysis": {
            "filter": {
                "map_poi_filter": {
                    "type": "synonym",
                    "synonyms": [
                    	"минфин,министертсво финансов",
                    	"минздрав,министерство здравоохранения",
                    	"минюст,министерство юстиции",
                    	"минтранспорта,министерство транспорта",
                    	"минобразования,министерство образования",
                    	"минкультуры,министерство культуры",
                    	"юракадемия,юридическая академия",
                    	"к-т,кинотеатр",
                    	"маг,магазин",
                    	"тц,торговый центр",
                    	"трк, торгово-развлекательный центр"
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
                    "street": {
                        "type": "string",
                        "analyzer": "simple"
                    },
                    "centroid": {
                        "type": "geo_point"
                    },
                    "name": {
                        "type": "string",
                        "analyzer": "simple"
                    }
                }
            }
        }
}`
