package main

import (
	"log"
	"os"

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
	if len(os.Args) > 1 && os.Args[1] == "web" {
		i.StartWebServer()
	}
	if err := i.Start(); err != nil {
		log.Fatal(err)
	}
	i.WaitStop()
	if err := i.Done(); err != nil {
		log.Fatal(err)
	}
}
