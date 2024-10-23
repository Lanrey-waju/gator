package main

import (
	"database/sql"
	"fmt"
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

	cmds := commands{commandHandlers: make(map[string]func(*state, command) error)}
	cmds.register("login", loginHandler)
	cmds.register("register", registerHandler)

	if err := cmds.run(&s, cmd); err != nil {
		log.Fatalf("Error running command: %s", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading updated config: %v", err)
	}
	fmt.Printf("DB URL: %s\nCurrent User: %s\n", cfg.DBUrl, cfg.CurrentUserName)
}
