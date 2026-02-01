package setup

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	logger := InitLogger()
	assert.NotNil(t, logger)
	assert.IsType(t, &slog.Logger{}, logger)
}

func TestLoadConfig(t *testing.T) {
	logger := slog.Default()
	cfg, err := LoadConfig(logger)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}
