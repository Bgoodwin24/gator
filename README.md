# **_Gator CLI_**
* Gator is a Go binary that functions as a RSS feed manager

## Required Software
* PostgreSQL
* Golang

## Installing Gator CLI
* `go install github.com/Bgoodwin24/gator`

## Setting up Gator Config File
* Manually create the config file in your home directory `~/.gatorconfig.json` 
    * `nano .gatorconfig.json` from ~/ (home dir)
    * Add this to the json file:
    `{
  "db_url": "postgres://example"
    }`

## Gator/Gator Command Usage:
* After installing, you can run gator from anywhere, as long as $HOME/go/bin is in your PATH
* Gator command syntax is as follows:
    * `gator <commandname> <commandparameters>`
* Ensure $HOME/go/bin is in your PATH if the gator command doesn’t work!

## User Management Commands:

* `gator register <username>` - Adds a new user to the database
* `gator login <username>` - Sets the current user in the config
* `gator users` - Lists users and indicates which one is currently logged in

## Feed Management Commands:
Requires authenticated user

* `gator addfeed <feedname> <feedurl>` - Adds a feed to a user
    * Example: `gator addfeed "Tech News" "https://example.com/feed.rss"`
* `gator follow <feedname> <feedurl>` - Follows a feed
* `gator following` - Lists all feeds the current user is following
* `gator unfollow <feedurl> <username>` - Unfollows a feed for the current user
    * Example: `gator unfollow "https://example.com/feed.rss" yourusername`
* `gator browse <optional limit>` - Shows posts from your feeds

## Utility Commands:
* `gator update` - Fetches posts from your feeds
* `gator reset` - Resets database for testing
* `gator agg <duration>` - Fetch RSS feeds and display the posts for the current user from the last `<duration>`, where duration is specified in seconds (e.g., 10s)
    * Example: `gator agg 10s`
