package utils

import (
	"os"
	"fmt"
)

func GetEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetPort() string {
	return GetEnv("PORT", "8080")
}

func GetHost() string {
	return GetEnv("HOST", "localhost")
}

func GetAddress() string {
	return fmt.Sprintf("%s:%s", GetHost(), GetPort())
}
