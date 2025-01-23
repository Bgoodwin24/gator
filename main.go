package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/Bgoodwin24/gator/internal/cli"
	"github.com/Bgoodwin24/gator/internal/config"
	"github.com/Bgoodwin24/gator/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'agg', 'login', 'register' command")
		os.Exit(1)
	}

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
	cmd.Register("reset", cli.HandlerReset)
	cmd.Register("users", cli.HandlerGetUsers)
	cmd.Register("agg", cli.HandlerAgg)
	cmd.Register("addfeed", cli.MiddlewareLoggedIn(cli.HandlerAddFeed))
	cmd.Register("feeds", cli.HandlerFeeds)
	cmd.Register("follow", cli.MiddlewareLoggedIn(cli.HandlerFollow))
	cmd.Register("following", cli.MiddlewareLoggedIn(cli.HandlerFollowing))
	cmd.Register("unfollow", cli.MiddlewareLoggedIn(cli.HandlerUnfollow))

	command := cli.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmd.Run(newState, command); err != nil {
		log.Fatal(err)
	}

}
