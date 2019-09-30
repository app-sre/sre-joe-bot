package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-joe/joe"
	joeslack "github.com/go-joe/slack-adapter"
	"github.com/machinebox/graphql"
	"github.com/nlopes/slack"
)

type SREBot struct {
	*joe.Bot
	Slack *slack.Client
	Gql   *graphql.Client
}

func main() {
	httpClient := http.Client{}
	g := graphql.NewClient(os.Getenv("APPINTERFACE_URL"), graphql.WithHTTPClient(&httpClient))

	s := slack.New(os.Getenv("SLACK_TOKEN"))
	sa := joeslack.Adapter(os.Getenv("SLACK_TOKEN"))
	b := &SREBot{
		Bot:   joe.New("sre-joe-bot", sa),
		Slack: s,
		Gql:   g,
	}

	adminIDS := strings.Split(os.Getenv("BOT_ADMIN_IDS"), ",")
	for _, userID := range adminIDS {
		userID = strings.TrimSpace(userID)
		b.Auth.Grant("admin", userID)
	}

	b.RespondRegex("^[Hh]i", b.CmdHi)
	b.RespondRegex("^[Hh]elp", b.CmdHelp)
	b.RespondRegex("^[Cc]luster [Ll]ist", b.CmdClusters)
	b.RespondRegex(".*", b.CmdInvalid)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}
