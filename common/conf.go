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
	err = json.Unmarshal(data, &C)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
