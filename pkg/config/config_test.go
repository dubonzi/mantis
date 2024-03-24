package config

import (
	"testing"

	"github.com/americanas-go/config"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaultConfig(t *testing.T) {
	SetDefaultConfig()
	config.Load()

	assert.Equal(t, "TEXT", config.String("log.format"))
	assert.Equal(t, "localhost:4318", config.String("otel.exporter.endpoint"))
}
