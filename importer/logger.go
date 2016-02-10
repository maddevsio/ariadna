package importer

import (
	log "github.com/mgutz/logxi/v1"
	"os"
)

var Logger log.Logger

func init() {
	Logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "importer")
}
