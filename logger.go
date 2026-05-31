package main

import (
	"github.com/phuslu/log"
)

func init() {
	log.DefaultLogger = log.Logger{
		Caller: 1,
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			QuoteString:    true,
			EndWithMessage: true,
		},
		Level: log.DebugLevel,
	}
}
