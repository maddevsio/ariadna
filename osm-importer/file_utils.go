package importer

import (
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
