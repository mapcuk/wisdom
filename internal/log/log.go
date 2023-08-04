package log

import (
	"log"

	"go.uber.org/zap"
)

var Glog *zap.Logger

func New(isDebug bool) *zap.Logger {
	conf := zap.NewProductionConfig()
	if isDebug {
		conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := conf.Build()
	if err != nil {
		log.Fatal(err)
	}
	Glog = logger
	return Glog
}

func Get() *zap.Logger {
	if Glog != nil {
		return Glog
	}
	return New(false)
}
