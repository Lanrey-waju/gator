package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/Lanrey-waju/gator.git/internal/config"
	"github.com/Lanrey-waju/gator.git/internal/database"

	_ "github.com/lib/pq"
)

func main() {

	arguments := os.Args
	if len(arguments) < 2 {
		log.Fatalf("Not enough arguments provided")
	}

	cmd := command{name: arguments[1], arg: arguments[2:]}

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatal("Error connecting to database")
	}

	dbQueries := database.New(db)

	s := state{db: dbQueries, cfg: &cfg}

	cmds := commands{registeredCommands: make(map[string]func(*state, command) error)}
	cmds.register("login", loginHandler)
	cmds.register("register", registerHandler)
	cmds.register("reset", resetHandler)
	cmds.register("users", getUsersHandler)
	cmds.register("agg", aggHandler)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", feedsHandler)
	cmds.register("follow", followHandler)
	cmds.register("following", followingHandler)

	if err := cmds.run(&s, cmd); err != nil {
		log.Fatalf("Error running command: %s", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading updated config: %v", err)
	}
}
