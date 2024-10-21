package main

import (
	"errors"
	"fmt"

	"github.com/Lanrey-waju/gator.git/internal/config"
)

type state struct {
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
	err := s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("%s has been set", username)
	return nil
}
