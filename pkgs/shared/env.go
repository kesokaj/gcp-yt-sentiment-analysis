package shared

import (
	"fmt" //
	"os"
	"strconv"
)

func GetEnvString(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {

		LogJSON("WARNING", fmt.Sprintf("Environment variable %s is not set. Using fallback value: %s", key, fallback), "")
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {

		LogJSON("WARNING", fmt.Sprintf("Environment variable %s is not set. Using fallback value: %d", key, fallback), "")
		return fallback
	}

	intValue, err := strconv.Atoi(valueStr)
	if err != nil {
		LogJSON("WARNING", fmt.Sprintf("Invalid integer value for environment variable %s: '%s'. Using fallback value: %d. Error: %v", key, valueStr, fallback, err), "")
		return fallback
	}
	return intValue
}
