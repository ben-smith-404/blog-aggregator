package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	var config Config
	filePath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func SetUser(userName string) error {
	config, err := Read()
	if err != nil {
		return err
	}
	config.CurrentUserName = userName
	err = write(config)
	if err != nil {
		return err
	}
	return nil
}

func write(config Config) error {
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path + "/" + configFileName, err
}
