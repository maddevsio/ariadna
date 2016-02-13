package common

import (
	log "github.com/gen1us2k/ariadna/logger"
)

var logger log.Logger

func init() {
	logger = log.L("common")
}
