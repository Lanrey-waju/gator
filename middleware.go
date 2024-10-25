package main

import (
	"context"
	"fmt"

	"github.com/Lanrey-waju/gator.git/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error retrieving user: %v\nAre you logged in?", err)
		}
		return handler(s, c, user)
	}
}
