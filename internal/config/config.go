package config

import (
	"fmt"
	"os"
)

type Config struct {
	Region        string
	ListenPort    string
	ListenAddress string
	ProjectID     string
}

func Load() (*Config, error) {
	region := getEnv("STACKIT_REGION", "eu01")
	port := getEnv("STACKIT_EXPORTER_LISTEN_PORT", "8080")
	projectID := os.Getenv("STACKIT_PROJECT_ID")

	if projectID == "" {
		return nil, fmt.Errorf("STACKIT_PROJECT_ID environment variable is required")
	}

	listenAddress := ":" + port

	return &Config{
		Region:        region,
		ListenPort:    port,
		ListenAddress: listenAddress,
		ProjectID:     projectID,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
