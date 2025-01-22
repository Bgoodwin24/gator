package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Bgoodwin24/gator/internal/config"
	"github.com/Bgoodwin24/gator/internal/database"
	"github.com/Bgoodwin24/gator/internal/rss"
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
		return fmt.Errorf("usage: %v <name>", cmd.Name)
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
	printUser(user)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}
	fmt.Println("Database reset successfully!")
	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get users: %w", err)
	}
	for i, user := range users {
		if i == 0 {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}
	fmt.Printf("Feed: %+v\n", feed)
	return nil
}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("expected 2 arguments: name and url")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't get users: %w", err)
	}

	feed, err := s.DB.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Println("Feed created successfully:")
	fmt.Println(feed)
	fmt.Println()
	fmt.Println("=====================================")

	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.FetchFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch feeds: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found")
		return nil
	}

	fmt.Printf("Found %d feeds:\n", len(feeds))

	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s, Feed URL: %s, User Name: %s\n", feed.FeedName, feed.FeedUrl, feed.UserName)
		fmt.Println("=====================================")
	}
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

func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:      %v\n", user.Name)
}
