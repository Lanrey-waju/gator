package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	file_path := home_dir + "/" + configFileName
	return file_path, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	dat, err := json.Marshal(c)
	if err != nil {
		return err
	}
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, dat, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Read() (Config, error) {
	conf := Config{}
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	dat, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(dat, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil

}
