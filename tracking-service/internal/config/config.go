package config

import (
	"fmt"
	"os"
	"strings"
	"wheres-my-pizza/tracking-service/internal/core/domain"
)

func GetEnv() *domain.DatabaseConfig {
	buf, err := os.ReadFile("./tracking-service/config.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result = strings.Split(string(buf), "\n")
	var table = make(map[string]map[string]string)
	var key string

	for i := range result {
		if strings.Contains(result[i], ":") {
			var result1 = strings.Fields(result[i])

			if len(result1) == 1 {
				key = strings.Trim(result1[0], ":")
				table[key] = make(map[string]string)
			} else if len(result1) == 2 {
				var key1 = strings.Trim(result1[0], ":")
				table[key][key1] = result1[1]
			}
		}
	}

	return NewDB(table)
}
func NewDB(table map[string]map[string]string) *domain.DatabaseConfig {
	db := &domain.DatabaseConfig{
		Host:     table["database"]["host"],
		Port:     table["database"]["port"],
		User:     table["database"]["user"],
		Password: table["database"]["password"],
		Database: table["database"]["database"],
	}

	return db
}
func Load() (*domain.DatabaseConfig, error) {

	return GetEnv(), nil
}
