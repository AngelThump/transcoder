package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ConfigStruct struct {
	StreamsAPI struct {
		Hostname string `json:"hostname"`
		AuthKey  string `json:"authKey"`
	} `json:"streamsApi"`
	Ingest struct {
		AuthKey string `json:"authKey"`
	} `json:"ingest"`
	DigitalOcean struct {
		Metadata struct {
			Hostname string `json:"hostname"`
		} `json:"metadata"`
	} `json:"digitalocean"`
	Cache struct {
		Hostname string `json:"hostname"`
	}
}

var Config *ConfigStruct

func NewConfig(configPath string) error {
	Config = &ConfigStruct{}

	d, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("unable to read config %v", err)
	}
	if err := json.Unmarshal(d, &Config); err != nil {
		log.Fatalf("unable to read config %v", err)
	}

	return nil
}

func ParseFlags() (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config.json", "path to config file")
	flag.Parse()

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
