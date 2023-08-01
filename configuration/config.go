package configuration

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	HOST   = "host"
	TOKEN  = "token"
	BROKER = "broker"
)

type Configuration struct {
	Host   string
	Token  string
	Broker string
}

func Load() *Configuration {
	viper.SetDefault(HOST, "localhost")
	viper.SetDefault(TOKEN, "")
	viper.SetDefault(BROKER, "localhost")
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")

	fileName := ".env"

	// Set the current directory as the starting point for the search
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Search for the file in folders above until found
	for {
		// Check if the file exists in the current directory
		if _, err := os.Stat(currentDir + "/" + fileName); err == nil {
			// Load the settings from the file
			viper.SetConfigFile(currentDir + "/" + fileName)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatal(err)
			}
			break
		}

		// Move to the parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root directory, break the loop
			break
		}

		// Update the current directory to the parent directory
		currentDir = parentDir
	}

	return &Configuration{
		Host:   viper.GetString(HOST),
		Token:  viper.GetString(TOKEN),
		Broker: viper.GetString(BROKER),
	}
}
