package main

import (
	"log"

	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/osm"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	i, err := osm.NewImporter(c)
	if err != nil {
		log.Fatal(err)
	}
	if err := i.Start(); err != nil {
		log.Fatal(err)
	}
	i.WaitStop()
}
