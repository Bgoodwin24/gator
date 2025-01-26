# **_Gator CLI_**
* Gator is a command-line utility written in Go that allows users to manage RSS feeds. With Gator, users can register accounts, follow or unfollow feeds, and update their collections—all powered by PostgreSQL for persistent storage and easy retrieval.

## Required Software
* PostgreSQL (Version 12 or higher recommended)
* Golang (Version 1.23.5 or higher)

## Installing Gator CLI
* `go install github.com/Bgoodwin24/gator`

## Setting up Gator Config File
1. Ensure PostgreSQL is installed and running on your machine.
2. Create a new database for `gator`. For example:
    `createdb gator`
3. Create the configuration file `~/.gatorconfig.json` in your home directory. You can do this using:
    `nano ~/.gatorconfig.json`
4. Add the following JSON configuration to the file:
    ```json
    {
      "db_url": "postgres://postgres:password@localhost:5432/gator"
    }
    ```

    Replace `password` with your actual PostgreSQL password (or leave it out if your PostgreSQL doesn't require one).

    ⚠️ **Note:** If you're using different credentials, a custom port, or a different database name, update the values accordingly.

## Gator/Gator Command Usage:
* After installing, you can run gator from anywhere by typing:
    * `gator <commandname> <commandparameters>`
* Ensure `$HOME/go/bin` is in your `PATH`. If the `gator` command doesn’t work:
    1. `echo $PATH`
    2. If not present, add it temporarily with:
       `export PATH=$PATH:$HOME/go/bin`
    3. To make it permanent, add the line above to your shell configuration file (e.g., `.bashrc` or `.zshrc`) and reload your shell:
       `source ~/.bashrc  # or source ~/.zshrc`
    4. Test that `PATH` update worked with: `gator --help`

## User Management Commands:
* `gator register <username>` - Adds a new user to the database
* `gator login <username>` - Sets the current user in the config
* `gator users` - Lists users and indicates which one is currently logged in

## Feed Management Commands:
Requires authenticated user

* `gator addfeed <feedname> <feedurl>` - Adds a feed to a user
    * Example: `gator addfeed "Example Name" "https://example.com/feed.rss"`
* `gator follow <feedname> <feedurl>` - Follows a feed
    * Example: `gator follow "Example Name" "https://example.com/feed.rss"`
* `gator following` - Lists all feeds the current user is following
* `gator unfollow <feedurl> <username>` - Unfollows a feed for the current user
    * Example: `gator unfollow "https://example.com/feed.rss" yourusername`
* `gator browse <optional limit>` - Shows posts from your feeds

## Utility Commands:
* `gator update` - Fetches posts from your feeds
    * Example: `gator update`
* `gator reset` - Resets database for testing 
    ⚠️**Note:** Use with caution. This resets your database and will remove all stored users and feeds.
    ⚠️ Ensure your `~/.gatorconfig.json` database URL is correct before running commands requiring database access.
    * Example: `gator reset`
* `gator agg <duration>` - Fetch RSS feeds and display the posts for the current user from the last `<duration>`, where duration is specified in seconds (e.g., `10s` means 10 seconds)
    * Example: `gator agg 10s`

## Example Workflow
### User Setup
* `gator register Examplename`
* `gator login Examplename`

#### Feed Management
**Add a feed, follow it, and update posts**
* `gator addfeed "Example Name" "https://example.com/feed.rss"`
* `gator follow "Example Name" "https://example.com/feed.rss"`
* `gator update`

**Browse and Manage Feeds**
* `gator browse`
* `gator following`
* `gator unfollow "https://example.com/feed.rss" examplename`
