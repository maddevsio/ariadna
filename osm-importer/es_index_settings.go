package importer

var ESSettings string = `
{"mappings":
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
					"fields": {
						"phonetic": {
							"type": "string",
							"analyzer": "phonetic"
						}
					},
					"type": "string"
				}
			}
		}
	},
	"settings":{
		"analysis":
			{
				"filter": {
					"my_metaphone": {
						"replace": false,
						"type": "phonetic",
						"encoder": "metaphone"
					},
					"ru_stemming": {
						"type": "snowball",
						"language": "Russian"
					}
				},
				"analyzer": {
					"phonetic": {
						"filter": ["standard", "lowercase", "my_metaphone"],
						"tokenizer": "standard"
					},
					"ru": {
						"filter": ["lowercase", "russian_morphology"],
						"type": "custom",
						"tokenizer": "standard"
					}
				}
			}
		}
	}`