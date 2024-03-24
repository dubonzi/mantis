package config

import (
	"github.com/americanas-go/config"
)

func SetDefaultConfig() {

	config.Add("server.port", 8080, "Server port")
	config.Add("server.disableStartupMessage", true, "Disable fiber startup message")

	config.Add("health.port", 8081, "Health endpoint port (must not be the same as the server port)")

	config.Add("otel.enabled", true, "Enable/disable Opentelemetry support")
	config.Add("otel.exporter.protocol", "http", "Protocol used to export Opentelemetry traces and metrics")
	config.Add("otel.exporter.endpoint", "localhost:4318", "Endpoint of the collecter for Opentelemetry traces and metrics")
	config.Add("otel.exporter.insecure", true, "Use insecure connection for Opentelemetry exporter")

	config.Add("loader.path.mapping", "files/mapping", "Path to the folder containing the mapping files")
	config.Add("loader.path.response", "files/response", "Path to the folder containing the response files")

	config.Add("log.level", "INFO", "Logging level")
	config.Add("log.format", "TEXT", "Logging format")

	config.Add("fx.log.enable", false, "Enable/disable fx startup log")
}
