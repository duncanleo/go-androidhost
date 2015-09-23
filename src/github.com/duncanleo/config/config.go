package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	SDKLocation string `json:"sdk-location"`
}

func GetConfig() (Config, error) {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return cfg, err
}