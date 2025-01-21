package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Bgoodwin24/gator/internal/config"
	"github.com/Bgoodwin24/gator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	DB     *database.Queries
	Config *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CommandNames map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("username required")
	}

	// Attempt to get user from database
	username := cmd.Args[0]
	_, err := s.DB.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user does not exist")
	}

	// User exists, set them as current user
	if err := s.Config.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}

	fmt.Printf("User name set: %s", cmd.Args[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("username required")
	}

	user, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	})

	if err != nil {
		return fmt.Errorf("user already exists")
	}

	// Set user in config file
	if err := s.Config.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}

	fmt.Printf("Created user: %+v\n", user)
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.CommandNames == nil {
		c.CommandNames = make(map[string]func(*State, Command) error)
	}
	c.CommandNames[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.CommandNames[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}
