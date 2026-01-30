package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/ben-smith-404/blog-aggregator/internal/config"
	"github.com/ben-smith-404/blog-aggregator/internal/database"
)

// a struct to hold the current state of the config file so we dont need to constantly
// look it up
type state struct {
	db  *database.Queries
	cfg *config.Config
}

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

func main() {
	var currentState state

	myConfig, err := config.Read()
	if err != nil {
		fmt.Printf("Error: %v/n", err)
		os.Exit(1)
	}

	currentState.cfg = &myConfig

	db, err := sql.Open("postgres", currentState.cfg.DbURL)
	if err != nil {
		fmt.Printf("Error: %v/n", err)
		os.Exit(1)
	}

	currentState.db = database.New(db)

	var commands commands = commands{command: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)

	inputs := os.Args
	if len(inputs) < 3 {
		fmt.Println("Error: not enough arguments were provided")
		os.Exit(1)
	}
	currentCommand := command{name: inputs[1], arguments: inputs[2:]}
	err = commands.run(&currentState, currentCommand)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// this is the helper function for the login command. it expects to be passed a command struct with
// a maximum of one string in the arguments slice. it will attempt to set the current user to the
// value of that string using the config in the state variable
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("login requires one username, %v were provided", len(cmd.arguments))
	}
	user, err := s.db.GetUser(context.Background(), sql.NullString{
		String: cmd.arguments[0],
		Valid:  true,
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name.String)
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

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("register requires one name, %v were provided", len(cmd.arguments))
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: sql.NullString{
			String: cmd.arguments[0],
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name.String)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v created with ID: %v with dates: %v\n", user.Name.String, user.ID, user.CreatedAt)
	return nil
}
