package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read(json_file string) (Config, error) {
	config := Config{}

	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("failed to locate home directory: %w", err)
	}

	//Create full file path
	path := homeDir + "/" + json_file

	//Open file
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	//Read file
	data, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	//Unmarshal JSON into config struct
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

const configFileName = ".gatorconfig.json"

// Sets username in Config struct
func (cfg *Config) SetUser(userName string) error {
	if len(userName) == 0 {
		return fmt.Errorf("userName cannot be empty")
	}

	cfg.CurrentUserName = userName
	return write(*cfg)
}

// Get full path to config file
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir + "/" + configFileName, nil
}

// Writes Config struct to the JSON file
func write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	//Overwrites existing file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	//Write JSON representation of cfg
	return encoder.Encode(cfg)
}
