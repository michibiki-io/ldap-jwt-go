package utility

import (
	"log"
	"os"
	"strconv"
)

var Log *Logger

type Logger struct{}

func (l *Logger) Debug(format string, args ...interface{}) {
	if os.Getenv("LOG_LELVE") == "DEBUG" {
		log.Printf(format, args...)
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if result, err := strconv.Atoi(value); err != nil {
			return fallback
		} else {
			return result
		}
	}
	return fallback
}

func GetBoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if result, err := strconv.ParseBool(value); err != nil {
			return fallback
		} else {
			return result
		}
	}
	return fallback
}
