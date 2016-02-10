package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func OpenFile(filename string) *os.File {
	// no file specified
	if len(filename) < 1 {
		Logger.Fatal("invalid file: you must specify a pbf path as arg[1]")
	}
	// try to open the file
	file, err := os.Open(filename)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	return file
}

func ReadConfig(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		Logger.Fatal(err.Error())
	}
	err = json.Unmarshal(data, &C)
	if err != nil {
		Logger.Fatal(err.Error())
	}
}
