package config

import (
	"encoding/json"
	"os"
)

// the config struct represents the structure of a json file stored in the users home directory
type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

// the name of the conifg file
const configFileName = ".gatorconfig.json"

// a public function allowing the config file to be read. Note that there is no logic to create
// the file if it does not exist yet. If the file does not exit, this will always throw an error,
// if the file does exist, a populated Config struct will be returned
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

// a public function to set the username in the config file
func (config Config) SetUser(userName string) error {
	config.CurrentUserName = userName
	err := write(config)
	if err != nil {
		return err
	}
	return nil
}

// private function allowing a new config to be written to the original file
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

// a private function tat returns the file path of the file
func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path + "/" + configFileName, err
}
