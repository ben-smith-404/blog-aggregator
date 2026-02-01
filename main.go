package main

import (
	"database/sql"
	"fmt"
	"os"

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

func main() {
	var currentState state

	myConfig, err := config.Read()
	checkError(err)

	currentState.cfg = &myConfig

	db, err := sql.Open("postgres", currentState.cfg.DbURL)
	checkError(err)

	currentState.db = database.New(db)
	commands := registerCommands()

	inputs := os.Args
	if len(inputs) < 2 {
		checkError(fmt.Errorf("Error: not enough arguments were provided"))
	}
	err = commands.run(&currentState, command{name: inputs[1], arguments: inputs[2:]})
	checkError((err))
}

// got sick of writing the error check every time
func checkError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
