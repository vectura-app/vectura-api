package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type CityConfig struct {
	ID  string `yaml:"id"`
	URL string `yaml:"url"`
}

type Config struct {
	SupportedCities []CityConfig `yaml:"cities"`
}

func LoadCitiesFromYAML(filename string) []CityConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	return config.SupportedCities
}
