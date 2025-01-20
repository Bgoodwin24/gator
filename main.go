package main

import (
	"log"
	"os"

	"github.com/Bgoodwin24/gator/internal/cli"
	"github.com/Bgoodwin24/gator/internal/config"
)

func main() {
	//Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	newState := &cli.State{Config: &cfg}

	cmd := cli.Commands{}

	cmd.Register("login", cli.HandlerLogin)
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}

	command := cli.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmd.Run(newState, command); err != nil {
		log.Fatal(err)
	}
}
