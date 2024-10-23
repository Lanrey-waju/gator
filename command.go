package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Lanrey-waju/gator.git/internal/config"
	"github.com/Lanrey-waju/gator.git/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	arg  []string
}

type commands struct {
	commandHandlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandHandlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commandHandlers[cmd.name]
	if !ok {
		return fmt.Errorf("command %s does not exist", cmd.name)
	}
	return handler(s, cmd)
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.arg) < 1 {
		return errors.New("gator expects at least one argument: the username")
	}
	username := cmd.arg[0]
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%s has been set", username)
	return nil
}

func registerHandler(s *state, cmd command) error {
	if len(cmd.arg) < 1 {
		return errors.New("gator expects at least one argument: the username")
	}
	username := cmd.arg[0]
	user, err := s.db.GetUser(context.Background(), username)
	if err == sql.ErrNoRows {
		user, err = s.db.CreateUser(context.Background(), database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		})
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		if err = s.cfg.SetUser(username); err != nil {
			return fmt.Errorf("error: %v", err)
		}
		fmt.Printf("%s was created. ID: %s, Created At: %v, Updated At: %v, name: %v", username, user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
		return nil
	} else if err != nil {
		return fmt.Errorf("Database error: %v", err)
	} else {
		return fmt.Errorf("user already exists")
	}

}
