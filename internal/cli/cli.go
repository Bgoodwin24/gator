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

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("expected 2 arguments: name and url")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

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

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
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

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("expected argument: url")
	}

	url := cmd.Args[0]

	current_user := s.Config.CurrentUserName
	if current_user == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	feeds, err := s.DB.FetchFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("feed not found with url: %s", url)
	}

	var feed database.FetchFeedsRow
	found := false
	for _, f := range feeds {
		if f.FeedUrl == url {
			feed = f
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("feed not found with url: %s", url)
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.FeedID,
	})
	if err != nil {
		return fmt.Errorf("faild to follow feed: %v", err)
	}

	fmt.Printf("User '%s' is now following feed '%s'.\n", user.Name, feed.FeedName)
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	followNames, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error fetching feed follows: %v", err)
	}

	if len(followNames) == 0 {
		fmt.Println("You are not following any feeds")
		return nil
	}

	for _, follow := range followNames {
		fmt.Println(follow.FeedName)
	}
	return nil
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	userID := user.ID

	_, err := s.DB.Unfollow(context.Background(), database.UnfollowParams{
		UserID: userID,
		Url:    cmd.Args[0],
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
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
