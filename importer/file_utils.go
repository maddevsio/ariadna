package importer

import (
	"os"
)

func OpenFile(filename string) *os.File {
	// no file specified
	if len(filename) < 1 {
		logger.Fatal("invalid file: you must specify a pbf path as arg[1]")
	}
	// try to open the file
	file, err := os.Open(filename)
	if err != nil {
		logger.Fatal(err.Error())
	}
	return file
}
