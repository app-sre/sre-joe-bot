package bot

import (
	"fmt"
	"net/http"

	"github.com/go-joe/joe"
	joeslack "github.com/go-joe/slack-adapter"

	"github.com/machinebox/graphql"
	"github.com/nlopes/slack"
)

type Bot struct {
	*joe.Bot
	Modules []joe.Module
	Slack   *slack.Client
	Gql     *graphql.Client
}

func WithSlackAdapter(token string) func(*Bot) error {
	return func(b *Bot) error {
		return b.setSlackClient(token)
	}
}

func (b *Bot) setSlackClient(token string) error {
	b.Slack = slack.New(token)
	b.Modules = append(b.Modules, joeslack.Adapter(token))

	return nil
}

func WithGraphqlClient(url string) func(*Bot) error {
	return func(b *Bot) error {
		return b.setGraphqlClient(url)
	}
}

func (b *Bot) setGraphqlClient(url string) error {
	b.Gql = graphql.NewClient(url, graphql.WithHTTPClient(&http.Client{}))
	return nil
}

func NewBot(name string, configs ...func(*Bot) error) (*Bot, error) {
	b := &Bot{}

	for _, config := range configs {
		err := config(b)
		if err != nil {
			return nil, fmt.Errorf("could not config adapter: %v", err)
		}
	}

	b.Bot = joe.New(name, b.Modules...)

	return b, nil
}
