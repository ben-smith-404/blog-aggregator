package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ben-smith-404/blog-aggregator/internal/config"
	"github.com/ben-smith-404/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

// a command represents an intruction the system can accept
type command struct {
	name      string
	arguments []string
}

// commands is the list of instructions, and the map to their functions
type commands struct {
	command map[string]func(*state, command) error
}

// the run method attepts to run the commands helper function, if there is no command of
// that name it returns an error, otherwise it runs the function in the map. If there is an
// error, it will return the error
func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.command[cmd.name]
	if !exists {
		return fmt.Errorf("there is no registered command with the name %v", cmd.name)
	}
	err := handler(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

// register adds a name: function pair to the commands struct
func (c *commands) register(name string, f func(*state, command) error) {
	c.command[name] = f
}

func registerCommands() commands {
	commands := commands{command: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)
	commands.register("follow", handlerFollow)
	commands.register("following", handlerFollowing)
	return commands
}

// this is the helper function for the login command. it expects to be passed a command struct with
// a maximum of one string in the arguments slice. it will attempt to set the current user to the
// value of that string using the config in the state variable
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("login requires one username, %v were provided", len(cmd.arguments))
	}
	user, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	*s.cfg, err = config.Read()
	if err != nil {
		return err
	}
	fmt.Printf("%v has been set as the current logged in user\n", s.cfg.CurrentUserName)
	return nil
}

// helper function for the register command. It takes a single parameter and creates them as a user
// in the database. If the user already exists it will throw an error. it then sets that user as the
// current user in the config file. It will print a message when it's successful
func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("register requires one name, %v were provided", len(cmd.arguments))
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v created with ID: %v with dates: %v\n", user.Name, user.ID, user.CreatedAt)
	return nil
}

// a very dangerous command to make testing easier. Reset truncates the user table in the database
func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("The users database table was reset")
	return nil
}

// the users function returns a list of users formatted as
// * user name
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Println("* " + user.Name + " (current)")
		} else {
			fmt.Println("* " + user.Name)
		}
	}
	return nil
}

// placeholder function to test the aggregate function
func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	fmt.Println(*feed)
	return nil
}

// add a feed to the database with a name, URL, and as the logged in user. It requres 2 parameters to be
// passed in, name and URL. It also creates a record that the logged in user is following a feed
func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("2 arguments expected, %v provided", len(cmd.arguments))
	}
	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	dbFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
		Url:       cmd.arguments[1],
		UserID:    currentUser.ID,
	})
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollower(context.Background(), database.CreateFeedFollowerParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    dbFeed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("New feed %v added. Followed by %v\n", dbFeed.Name, currentUser.Name)
	return nil
}

// prints a list of feeds and the name of the user who created each feed
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsAndUserName(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("Feed: %v with URL: %v was created by: %v\n", feed.Name, feed.Url, feed.UserName)
	}
	return nil
}

// this command takes a single input, a URL and subscribes the user to the feed. If the URL does not exist
// a new feed will not be created
func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("1 argument expected, %v provided", len(cmd.arguments))
	}
	feed, err := s.db.GetFeedsByURL(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}
	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	feedFollower, err := s.db.CreateFeedFollower(context.Background(), database.CreateFeedFollowerParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	})
	fmt.Printf("%v is followed by %v\n", feedFollower.FeedName, feedFollower.UserName)
	return nil
}

// this command prints a list of all the feeds the user is currently following
func handlerFollowing(s *state, cmd command) error {
	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return nil
	}
	feeds, err := s.db.GetFeedsUserFollows(context.Background(), currentUser.ID)
	if err != nil {
		return nil
	}
	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
	}
	return nil
}
