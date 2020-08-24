package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-joe/joe"
	joeslack "github.com/go-joe/slack-adapter/v2"
	"go.uber.org/zap"

	"github.com/machinebox/graphql"
	"github.com/slack-go/slack"
)

type Bot struct {
	*joe.Bot
	Modules      []joe.Module
	Commands     []BotCommand
	Slack        *slack.Client
	Gql          *graphql.Client
	GqlBasicAuth string
	Version      string
}

type BotCommand struct {
	Command string
	Fun     func(joe.Message) error
}

type ResponseFunc func(joe.Message) error

func WithSlackAdapter(token string) func(*Bot) error {
	return func(b *Bot) error {
		return b.setSlackClient(token)
	}
}

func (b *Bot) setSlackClient(token string) error {
	b.Slack = slack.New(token)

	_, err := b.Slack.AuthTest()
	if err != nil {
		return fmt.Errorf("could not connect to slack: %v", err)
	}

	b.Modules = append(b.Modules, joeslack.Adapter(token))
	return nil
}

func WithGraphqlClient(url, basicAuth string) func(*Bot) error {
	return func(b *Bot) error {
		return b.setGraphqlClient(url, basicAuth)
	}
}

func (b *Bot) setGraphqlClient(url, basicAuth string) error {
	b.Gql = graphql.NewClient(url, graphql.WithHTTPClient(&http.Client{}))
	b.GqlBasicAuth = basicAuth
	return b.testGraphqlClient()
}

func WithVersionCommand(version string) func(*Bot) error {
	return func(b *Bot) error {
		return b.setVersionCommand(version)
	}
}

func (b *Bot) setVersionCommand(version string) error {
	b.Version = version
	b.Commands = append(b.Commands, BotCommand{
		Command: "version",
		Fun: func(msg joe.Message) error {
			msg.Respond("This bot is running version %s", b.Version)
			return nil
		},
	})
	return nil
}

func (b *Bot) testGraphqlClient() error {
	request := graphql.NewRequest(`{
		__schema {
			types {
				name
			}
		}
	}`)
	request.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

	var response struct{}
	if err := b.Gql.Run(context.Background(), request, &response); err != nil {
		return err
	}
	return nil
}

func (b *Bot) Log(fun func(joe.Message) error) func(joe.Message) error {
	return ResponseFunc(func(msg joe.Message) error {
		b.Logger.Info("Command", zap.String("user", msg.AuthorID), zap.String("command", msg.Text))
		return fun(msg)
	})
}

func (b *Bot) Authenticate(scope string, fun func(joe.Message) error) func(joe.Message) error {
	return ResponseFunc(func(msg joe.Message) error {
		err := b.Auth.CheckPermission(scope, msg.AuthorID)
		if err != nil {
			return msg.RespondE("You are not allowed to run this command")
		}
		return fun(msg)
	})
}

func NewBot(name string, configs ...func(*Bot) error) (*Bot, error) {
	b := &Bot{}

	for _, config := range configs {
		err := config(b)
		if err != nil {
			return nil, err
		}
	}

	b.Bot = joe.New(name, b.Modules...)

	for _, c := range b.Commands {
		b.Respond(c.Command, c.Fun)
	}

	return b, nil
}
