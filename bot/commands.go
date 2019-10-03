package bot

import (
	"context"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/go-joe/joe"
	"github.com/machinebox/graphql"
	"go.uber.org/zap"
)

func (b *Bot) CmdInvalid(msg joe.Message) error {
	// TODO: enable this after fix is merged https://github.com/go-joe/slack-adapter/pull/6
	//b.Logger.Info("Invalid Command", zap.String("user", msg.AuthorID), zap.String("command", msg.Text))
	//msg.Respond("Hi! I'm a bot and `%s` in an invalid command. Try `help`", msg.Text)
	return nil
}

func (b *Bot) CmdHelp(msg joe.Message) error {
	b.Logger.Info("Command", zap.String("user", msg.AuthorID), zap.String("command", msg.Text))
	var resp = []string{
		"Here's how you can interract with me",
		"------------------------------------",
		"cluster list: List known clusters",
	}
	msg.Respond(pre(resp))
	return nil
}

func (b *Bot) CmdHi(msg joe.Message) error {
	user, err := b.Slack.GetUserInfo(msg.AuthorID)
	if err != nil {
		return err
	}
	msg.Respond("Hey %s, how's it going?", user.Profile.DisplayName)
	return nil
}

const GqlClusters = `
{
	cluster: clusters_v1 {
		name
	}
}
`

func (b *Bot) CmdClusters(msg joe.Message) error {
	err := b.Auth.CheckPermission("admin", msg.AuthorID)
	if err != nil {
		return msg.RespondE("You are not allowed to run this command")
	}

	req := graphql.NewRequest(GqlClusters)
	req.Header.Set("Authorization", "Basic "+os.Getenv("APPINTERFACE_AUTH"))

	var res struct {
		Cluster []struct {
			Name string
		}
	}
	if err := b.Gql.Run(context.Background(), req, &res); err != nil {
		log.Fatal(err)
	}

	var clusters []string
	for _, cluster := range res.Cluster {
		clusters = append(clusters, strings.TrimSpace(cluster.Name))
	}
	sort.Strings(clusters)

	msg.Respond(pre(clusters))

	return nil
}
