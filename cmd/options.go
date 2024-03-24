package main

import (
	"github.com/americanas-go/config"
	igzap "github.com/americanas-go/log/contrib/go.uber.org/zap.v1"
)

func zapOptions() *igzap.Options {
	return &igzap.Options{
		Console: struct {
			Enabled   bool
			Level     string
			Formatter string
		}{
			Enabled:   true,
			Level:     config.String("log.level"),
			Formatter: config.String("log.format"),
		},
	}
}
