package main

import (
	"fmt"

	"github.com/Bgoodwin24/gator/internal/config"
)

func main() {
	//Read the config file
	cfg, err := config.Read(".gatorconfig.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	//Set the current user and persist to the file
	err = cfg.SetUser("Billy")
	if err != nil {
		fmt.Println("Error setting user:", err)
		return
	}

	//Read the config file again to verify changes
	cfg, err = config.Read(".gatorconfig.json")
	if err != nil {
		fmt.Println("Error reading updated config:", err)
		return
	}

	//Print the updated config
	fmt.Println("Database URL:", cfg.DBUrl)
	fmt.Println("Current User:", cfg.CurrentUserName)
}
