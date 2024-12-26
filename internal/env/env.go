package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetString(key, fallback string) string {
	if val, found := lookupEnv(key); found {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val, found := lookupEnv(key); found {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if val, found := lookupEnv(key); found {
		if boolVal, err := strconv.ParseBool(strings.ToLower(val)); err == nil {
			return boolVal
		}
	}
	return fallback
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	if val, found := lookupEnv(key); found {
		if durationVal, err := time.ParseDuration(strings.ToLower(val)); err == nil {
			return durationVal
		}
	}
	return fallback
}

func lookupEnv(key string) (string, bool) {
	val := cleanString(os.Getenv(cleanString(key)))
	return val, len(val) > 0
}

func cleanString(str string) string {
	return strings.TrimSpace(strings.Trim(str, "\""))
}
