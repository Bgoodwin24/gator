package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
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

func ScrapeFeeds(s *State) {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("couldn't get next feeds to fetch", err)
		return
	}
	log.Println("Found a feed to fetch!")
	ScrapeFeed(s.DB, feed)
}

func ScrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("couldn't collect feed %s fetched: %v:", feed.Name, err)
		return
	}

	for _, item := range feedData.Channel.Item {
		fmt.Printf("Found post: %s\n", item.Title)

		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// Keep the fallbacks in case other feeds use different formats
			publishedAt, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				publishedAt, err = time.Parse(time.RFC822, item.PubDate)
				if err != nil {
					log.Printf("couldn't parse published at with any format: %v", err)
					continue
				}
			}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Error saving post: %v", err)
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		ScrapeFeeds(s)
	}
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
		return fmt.Errorf("expected argument: feed name")
	}

	feedName := cmd.Args[0]

	feeds, err := s.DB.FetchFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching feeds: %w", err)
	}

	var feed database.FetchFeedsRow
	found := false
	for _, f := range feeds {
		if f.FeedName == feedName {
			feed = f
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("feed not found with name: %s", feedName)
	}

	_, err = s.DB.GetFeedFollow(context.Background(), database.GetFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.FeedID,
	})
	if err == nil {
		return fmt.Errorf("you are already following '%s'", feed.FeedName)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking feed follow: %w", err)
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.FeedID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}

	fmt.Printf("Following '%s'\n", feed.FeedName)
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
	if len(cmd.Args) < 1 {
		return fmt.Errorf("feed name is required")
	}

	feedName := cmd.Args[0]
	userID := user.ID

	err := s.DB.Unfollow(context.Background(), database.UnfollowParams{
		UserID: userID,
		Name:   feedName,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}
	return nil
}

func HandlerUpdate(s *State, cmd Command, user database.User) error {
	ScrapeFeeds(s)
	return nil
}

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2

	if len(cmd.Args) > 0 {
		userLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
		limit = userLimit
	}

	posts, err := s.DB.GetPostsForUsers(context.Background(), database.GetPostsForUsersParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting posts: %v", err)
	}

	for i, post := range posts {
		if i > 0 {
			fmt.Println("----------------------------------------")
		}
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Url: %s\n", post.Url)
		fmt.Printf("Published: %s\n", post.PublishedAt.Format("2006-01-02 15:04:05 -0700"))
		fmt.Printf("Feed: %s\n", post.FeedName)
		if post.Description != "" {
			if !strings.Contains(post.Description, "<p>Article URL:") {
				cleanDesc := stripHTML(post.Description)
				if cleanDesc != "" {
					fmt.Printf("Description: %s\n", cleanDesc)
				}
			}
		}
		fmt.Println()
	}
	return nil
}

func stripHTML(html string) string {
	text := strings.ReplaceAll(html, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")

	var buf strings.Builder
	inTag := false
	for _, r := range text {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}
	result := strings.TrimSpace(buf.String())
	result = strings.Join(strings.Fields(result), " ")

	return result
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
