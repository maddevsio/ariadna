package importer

import (
	log "github.com/gen1us2k/ariadna/logger"
)

var Logger log.Logger

func init() {
	Logger = log.L("importer")
}
