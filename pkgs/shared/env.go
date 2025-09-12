package shared

import (
	"os"
	"strconv"
)

func GetEnvString(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		Logger.Warn("Environment variable not set, using fallback.", "key", key, "fallback", fallback)
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {
		Logger.Warn("Environment variable not set, using fallback.", "key", key, "fallback", fallback)
		return fallback
	}

	intValue, err := strconv.Atoi(valueStr)
	if err != nil {
		Logger.Warn("Invalid integer value for environment variable, using fallback.", "key", key, "value", valueStr, "fallback", fallback, "error", err)
		return fallback
	}
	return intValue
}
