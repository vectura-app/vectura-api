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

// Helper function to load config from YAML
func loadConfig() (*Config, error) {
	filename := "/data/cities.yaml"
	data, err := os.ReadFile(filename)
	if err != nil {
		// Try fallback location
		filename = "cities.yaml"
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read cities.yaml from both locations: %w", err)
		}
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

func LoadCitiesFromYAML() []CityConfig {
	config, err := loadConfig()
	gostfu(err)
	return config.SupportedCities
}

func GetCityIDIndex() []string {
	config, err := loadConfig()
	gostfu(err)

	var idx []string
	for _, city := range config.SupportedCities {
		idx = append(idx, city.ID)
	}
	return idx
}

func SaveGTFS(url string, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
