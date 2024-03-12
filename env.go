package main

import (
	"os"
	"strings"
)

func getEnv(key string) string {
	envFile := ".env"
	stream, err := os.ReadFile(envFile)
	check(err)
	envVars := strings.Split(string(stream), "\n")
	for _, n := range envVars {
		if strings.HasPrefix(n, key) {
			value := strings.TrimPrefix(n, key+"=")
			return strings.Trim(value, "\"")
		}
	}

	return ""
}
