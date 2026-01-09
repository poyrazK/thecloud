package setup

import (
	"log/slog"
	"os"

	"github.com/poyrazk/thecloud/internal/platform"
)

// InitLogger initializes and returns the application logger
func InitLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}

// LoadConfig loads the application configuration
func LoadConfig(logger *slog.Logger) (*platform.Config, error) {
	cfg, err := platform.NewConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return nil, err
	}
	return cfg, nil
}
