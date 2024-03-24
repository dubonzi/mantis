package main

import (
	"github.com/americanas-go/config"
	igzap "github.com/americanas-go/log/contrib/go.uber.org/zap.v1"
)

func loadDefaultConfig() {

	config.Add("server.port", 8080, "Server port")
	config.Add("server.disableStartupMessage", false, "Disable fiber startup message")

	config.Add("health.port", 8081, "Health endpoint port (must not be the same as the server port)")

	config.Add("loader.path.mapping", "files/mapping", "Path to the folder containing the mapping files")
	config.Add("loader.path.response", "files/response", "Path to the folder containing the response files")

	config.Add("log.level", "INFO", "Logging level")
	config.Add("log.format", "TEXT", "Logging format")

	config.Add("fx.log.enable", false, "Enable/disable fx startup log")

}

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
