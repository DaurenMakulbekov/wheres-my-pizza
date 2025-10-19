package config

import (
	"fmt"
	"os"
	"strings"
)

type DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RabbitMQ struct {
	Host     string
	Port     string
	User     string
	Password string
}

type AppConfig struct {
	DB       *DB
	RabbitMQ *RabbitMQ
}

func GetEnv() map[string]map[string]string {
	buf, err := os.ReadFile("./order-service/configs/config.yaml")
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

	return table
}

func NewDB(table map[string]map[string]string) *DB {
	db := &DB{
		Host:     table["database"]["host"],
		Port:     table["database"]["port"],
		User:     table["database"]["user"],
		Password: table["database"]["password"],
		Name:     table["database"]["database"],
	}

	return db
}

func NewRabbitMQ(table map[string]map[string]string) *RabbitMQ {
	config := &RabbitMQ{
		Host:     table["rabbitmq"]["host"],
		Port:     table["rabbitmq"]["port"],
		User:     table["rabbitmq"]["user"],
		Password: table["rabbitmq"]["password"],
	}

	return config
}

func NewAppConfig() *AppConfig {
	table := GetEnv()

	config := &AppConfig{
		DB:       NewDB(table),
		RabbitMQ: NewRabbitMQ(table),
	}

	return config
}
