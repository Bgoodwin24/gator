package cli

import (
	"fmt"

	"github.com/Bgoodwin24/gator/internal/config"
)

type State struct {
	*config.Config
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

	if err := s.Config.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}

	fmt.Printf("User name set: %s", cmd.Args[0])
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
