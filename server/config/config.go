// Package config TestBert Configuration
package config

import (
	"log"
	"os"
)

type Configuration struct {
	ServerPort      string
	AuthSecret      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	OtlpEndpoint    string
	OtlpPort        string
	OtelServiceName string
	OtelEnvironment string
}

func NewConfig() *Configuration {
	cfg := &Configuration{
		ServerPort:      os.Getenv("TESTBERT_SERVER_PORT"),
		AuthSecret:      os.Getenv("TESTBERT_AUTH_SECRET"),
		DBHost:          os.Getenv("TESTBERT_DB_HOST"),
		DBPort:          os.Getenv("TESTBERT_DB_PORT"),
		DBUser:          os.Getenv("TESTBERT_DB_USER"),
		DBPassword:      os.Getenv("TESTBERT_DB_PASSWORD"),
		DBName:          os.Getenv("TESTBERT_DB_NAME"),
		OtlpEndpoint:    os.Getenv("TESTBERT_OTLP_ENDPOINT"),
		OtlpPort:        os.Getenv("TESTBERT_OTLP_PORT"),
		OtelServiceName: os.Getenv("TESTBERT_OTEL_SERVICE_NAME"),
		OtelEnvironment: os.Getenv("TESTBERT_OTEL_ENVIRONMENT"),
	}

	if cfg.AuthSecret == "" {
		log.Fatal("TESTBERT_AUTH_SECRET not set")
	}
	if cfg.DBPassword == "" {
		log.Fatal("TESTBERT_DB_PASSWORD not set")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "50013"
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "5433"
	}
	if cfg.DBUser == "" {
		cfg.DBUser = "testy"
	}
	if cfg.DBName == "" {
		cfg.DBName = "testbert"
	}
	if cfg.OtlpEndpoint == "" {
		cfg.OtlpEndpoint = "localhost"
	}
	if cfg.OtlpPort == "" {
		cfg.OtlpPort = "4317"
	}
	if cfg.OtelServiceName == "" {
		cfg.OtelServiceName = "testbert"
	}
	if cfg.OtelEnvironment == "" {
		cfg.OtelEnvironment = "development"
	}

	return cfg
}
