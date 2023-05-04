package config

import (
	"encoding/json"
	"go-empty/pkg/db"
	"io/ioutil"
)

type Config struct {
	Server Server    `json:"Server"`
	DB     db.Config `json:"DB"`
}

type Server struct {
	Port string `json:"Port"`
	IP   string `json:"IP"`
}

func New(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
