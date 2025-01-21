package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/Bgoodwin24/gator/internal/cli"
	"github.com/Bgoodwin24/gator/internal/config"
	"github.com/Bgoodwin24/gator/internal/database"
)

func main() {
	// Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	newState := &cli.State{
		DB:     dbQueries,
		Config: &cfg,
	}

	cmd := cli.Commands{}
	cmd.Register("login", cli.HandlerLogin)
	cmd.Register("register", cli.HandlerRegister)

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
