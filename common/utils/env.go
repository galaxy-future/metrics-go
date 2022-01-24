package utils

import (
	"github.com/galaxy-future/metrics-go/common/logger"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

func TryGetStringEnvironment(name, defaultValue string) string {
	content := os.Getenv(name)
	valueFromEnv := os.Getenv(name)
	if valueFromEnv == "" {
		logger.GetLogger().Warn("no environment configured, default value used", zap.String("environment", name), zap.String("default", defaultValue))
		return defaultValue
	}
	return content
}
func TryGetIntEnvironment(name string, defaultValue int) int {
	content := os.Getenv(name)
	valueFromEnv := os.Getenv(name)
	if valueFromEnv == "" {
		logger.GetLogger().Warn("no environment configured, default value used", zap.String("environment", name), zap.Int("default", defaultValue))
		return defaultValue
	}

	value, err := strconv.Atoi(content)
	if err != nil {
		logger.GetLogger().Error("environment configured error, default value used", zap.String("environment", name), zap.Int("default", defaultValue))
		return defaultValue
	}
	return value
}
func TryGetDurationEnvironment(name string, defaultValue time.Duration) time.Duration {
	content := os.Getenv(name)
	valueFromEnv := os.Getenv(name)
	if valueFromEnv == "" {
		logger.GetLogger().Warn("no environment configured, default value used", zap.String("environment", name), zap.Duration("default", defaultValue))
		return defaultValue
	}

	value, err := time.ParseDuration(content)
	if err != nil {
		logger.GetLogger().Error("environment configured error, default value used", zap.String("environment", name), zap.Duration("default", defaultValue))
		return defaultValue
	}
	return value
}
