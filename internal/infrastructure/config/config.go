package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type EarthquakeConfig struct {
	BrokerURL     string `required:"true" split_words:"true"`
	CkanURL       string `envconfig:"CKAN_URL"`
	CkanDatastore string `envconfig:"CKAN_DATASTORE_ID"`
	CkanKey       string `envconfig:"CKAN_API_KEY"`
}

func (s EarthquakeConfig) String() string {
	return fmt.Sprintf(`
		BrokerURL: %s,
		CkanURL: %s,
		CkanDatastore: %s,
		CkanKey: beginning with %s,
	`, s.BrokerURL, s.CkanURL, s.CkanDatastore, s.CkanKey[:5],
	)
}

func LoadEarthquakeConfig() (*EarthquakeConfig, error) {
	err := godotenv.Load(".env.example")
	//err := godotenv.Load() TODO use this on production

	if err != nil {
		log.Printf("could not load configuration from .env file: %v", err)
	}
	var c EarthquakeConfig
	err = envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	log.Printf("Loaded configuration:%+s", c)
	return &c, nil
}
