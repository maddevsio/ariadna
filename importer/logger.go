package importer

import (
	log "github.com/maddevsio/ariadna/logger"
)

var Logger log.Logger

func init() {
	Logger = log.L("importer")
}
