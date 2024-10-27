package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	user, err := s.db.GetUserByName(context.Background(), username)
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
	user, err := s.db.GetUserByName(context.Background(), username)
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
	time_between_reqs := cmd.arg[0]
	timeBetweenRequests, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("error parsing aggregation interval: %v", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			return fmt.Errorf("error aggregating feeds: %v", err)
		}
	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 2 {
		return fmt.Errorf("command expects two arguments: name and url")
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
	cmd = command{name: "follow", arg: []string{feed.Url}}
	if err := followHandler(s, cmd, user); err != nil {
		return fmt.Errorf("error following feed %s: %v", feed.Url, err)
	}
	fmt.Printf("Feed Name: %v\n Feed Url: %v\n", feed.Name, feed.Url)
	return nil
}

func feedsHandler(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feeds: %v", err)
	}
	for _, feed := range feeds {

		fmt.Printf("Feed Name: %v\nFeed URL: %v\nCreator: %v\n", feed.FeedName, feed.Url, feed.Creator)
	}
	return nil
}

func followHandler(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 1 {
		return fmt.Errorf("follow command requires one argument: url")
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arg[0])
	if err != nil {
		return fmt.Errorf("error retrieving feed with url: %v", err)
	}
	feed_follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating a feed follow: %v", err)
	}
	fmt.Printf("Follower: %v\n", feed_follow.Follower)
	fmt.Printf("Feed: %v\n", feed_follow.Following)
	return nil
}

func followingHandler(s *state, cmd command, user database.User) error {
	feed_follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error retrieving feed follows for user %s: %v", user.Name, err)
	}
	for _, feed_follow := range feed_follows {
		fmt.Println(feed_follow.Follower)
		fmt.Println(feed_follow.Feed, "\n")
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 1 {
		return fmt.Errorf("follow command requires one argument: url")
	}
	url := cmd.arg[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error retrieving feed with url %s: %v", url, err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %v", err)
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	var err error
	if len(cmd.arg) == 1 {
		if limit, err = strconv.Atoi(cmd.arg[0]); err != nil {
			return fmt.Errorf("error converting limit argument: %v", err)
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error retrieving posts for user %v", user.Name)
	}
	for _, post := range posts {
		fmt.Println("Post Title:", post.Title)
		fmt.Println("Post Description:", post.Description)
		fmt.Println("")
	}
	return nil
}
