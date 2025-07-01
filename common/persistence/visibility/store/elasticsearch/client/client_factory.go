package client

import (
	"fmt"
	"net/http"

	"go.temporal.io/server/common/log"
)

func NewClient(config *Config, httpClient *http.Client, logger log.Logger) (Client, error) {
	switch config.Version {
	case "v8", "v7", "":
		return newClient(config, httpClient, logger)
	default:
		return nil, fmt.Errorf("not supported Elasticsearch version: %v", config.Version)
	}
}

func NewCLIClient(config *Config, logger log.Logger) (CLIClient, error) {
	switch config.Version {
	case "v8", "v7", "":
		return newClient(config, nil, logger)
	default:
		return nil, fmt.Errorf("not supported Elasticsearch version: %v", config.Version)
	}
}

func NewFunctionalTestsClient(config *Config, logger log.Logger) (IntegrationTestsClient, error) {
	switch config.Version {
	case "v8", "v7", "":
		return newClient(config, nil, logger)
	default:
		return nil, fmt.Errorf("not supported Elasticsearch version: %v", config.Version)
	}
}

func NewElasticClient(config *Config, httpClient *http.Client, logger log.Logger) (NewEsClient, error) {
	switch config.Version {
	case "v8", "official", "go-elasticsearch":
		return NewESClient(config, httpClient, logger)
	case "v7", "":
		return nil, fmt.Errorf("v7 not implemented yet")
	default:
		return nil, fmt.Errorf("not supported version: %v", config.Version)
	}
}