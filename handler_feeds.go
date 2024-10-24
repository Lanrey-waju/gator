package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Lanrey-waju/gator.git/internal/database"
	"github.com/google/uuid"
)

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
