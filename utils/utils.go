package utils

import (
	"fmt"
	"io"
	"net/http"
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

func gostfu(err error) {
	if err != nil {
		panic(err)
	}
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

func FetchGTFS(url string) ([]byte, error) {
	resp, err := http.Get(url)

	gostfu(err)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)

	gostfu(err)

	return data, nil
}
