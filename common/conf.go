package common

import (
	"encoding/json"
	"io/ioutil"
)

func ReadConfig(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = json.Unmarshal(data, &IC)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
