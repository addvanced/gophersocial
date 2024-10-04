package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetString(key, fallback string) string {
	if val, found := os.LookupEnv(key); found {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val, found := os.LookupEnv(key); found {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if val, found := os.LookupEnv(key); found {
		if boolVal, err := strconv.ParseBool(strings.TrimSpace(val)); err == nil {
			return boolVal
		}
	}
	return fallback
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	if val, found := os.LookupEnv(key); found {
		if durationVal, err := time.ParseDuration(val); err == nil {
			return durationVal
		}
	}
	return fallback
}
