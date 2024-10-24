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
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.registeredCommands[cmd.name]
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
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
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

func resetHandler(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("error resetting database: %v", err)
	}
	if err := s.cfg.SetUser("Nil"); err != nil {
		return fmt.Errorf("Error setting user to NIL: %v\n", err)
	}
	fmt.Printf("Reset successful!\n")
	return nil

}

func getUsersHandler(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving users: %v\n", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func aggHandler(s *state, cmd command) error {
	rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("Error fetching feed: %v", err)
	}
	fmt.Printf("Channel Title: %s\n", rssFeed.Channel.Title)
	for _, feed := range rssFeed.Channel.Item {
		fmt.Printf("Title: %s\n", feed.Title)
		fmt.Printf("Description: %s\n\n", feed.Description)
	}
	return nil

}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arg) < 2 {
		return fmt.Errorf("command expects two arguments: name and url")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Error retrieving user: %v", err)
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      cmd.arg[0],
		Url:       cmd.arg[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed: %v", err)
	}
	fmt.Printf("Feed Name: %v\n Feed Url: %v\n", feed.Name, feed.Url)
	return nil
}
