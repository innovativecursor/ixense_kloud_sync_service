package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Env string `yaml:"env"`

	Database struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		DatabaseName string `yaml:"databaseName"`
	} `yaml:"database"`
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
	ERP struct {
		BaseURL string `yaml:"baseUrl"`
		APIKey  string `yaml:"apiKey"`
	} `yaml:"erp"`
	KloudPX struct {
		BaseURL    string `yaml:"baseUrl"`
		ServiceKey string `yaml:"serviceKey"`
	} `yaml:"kloudpx"`

	Internal struct {
		ServiceKey string `yaml:"serviceKey"`
	} `yaml:"internal"`
}

func Env() (Config, error) {
	var config Config

	home := os.Getenv("ENV_PATH")
	filePath := filepath.Join(home, "env.yaml")

	envData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(envData, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
