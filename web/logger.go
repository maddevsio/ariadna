package web

import (
	log "github.com/maddevsio/ariadna/logger"
)

var logger log.Logger

func init() {
	logger = log.L("web")
}
