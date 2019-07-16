package config

import "github.com/spf13/viper"

type Ariadna struct {
	ElasticIndex  string   `json:"elastic_index" mapstructure:"elastic_index"`
	ElasticURLs   []string `json:"elastic_urls" mapstructure:"elastic_urls"`
	OSMFilename   string   `json:"osm_filename" mapstructure:"osm_filename"`
	IndexSettings string   `json:"index_settings" mapstructure:"index_settings"`
	OSMURL        string   `json:"osm_url" mapstructure:"osm_url"`
}

func Get() (*Ariadna, error) {
	var a Ariadna
	viper.SetConfigName("ariadna")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	envVariables := []string{"elastic_index", "elastic_urls"}
	for _, env := range envVariables {
		if err := viper.BindEnv(env); err != nil {
			return nil, err
		}
	}
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
