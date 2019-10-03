package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		"hi: Say hi!",
		"help: you're reading it",
		"version: reports the bot version",
		"",
		"get clusters: List known clusters",
		"",
		"get schema: List app-interface schemas",
		"get schema <schema>: Show app-interface schema",
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

func (b *Bot) CmdGetClusters(msg joe.Message) error {
	err := b.Auth.CheckPermission("admin", msg.AuthorID)
	if err != nil {
		return msg.RespondE("You are not allowed to run this command")
	}

	req := graphql.NewRequest(`{
		cluster: clusters_v1 {
			name
		}
	}`)
	req.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

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

func (b *Bot) CmdGetSchemas(msg joe.Message) error {
	err := b.Auth.CheckPermission("admin", msg.AuthorID)
	if err != nil {
		return msg.RespondE("You are not allowed to run this command")
	}

	req := graphql.NewRequest(`{
		__schema {
			types {
				name
			}
		}
	}`)
	req.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

	var res struct {
		Schema struct {
			Types []struct {
				Name string `json:"name"`
			} `json:"types"`
		} `json:"__schema"`
	}
	if err := b.Gql.Run(context.Background(), req, &res); err != nil {
		log.Fatal(err)
	}

	var schemaTypes []string
	for _, schemaType := range res.Schema.Types {
		schemaTypes = append(schemaTypes, strings.TrimSpace(schemaType.Name))
	}
	sort.Strings(schemaTypes)

	msg.Respond(pre(schemaTypes))

	return nil
}

func (b *Bot) CmdGetSchema(msg joe.Message) error {
	err := b.Auth.CheckPermission("admin", msg.AuthorID)
	if err != nil {
		return msg.RespondE("You are not allowed to run this command")
	}

	req := graphql.NewRequest(fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
					ofType {
						name
						kind
					}
				}
			}
		}
	}`, msg.Matches[0]))
	req.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

	var res struct {
		Type struct {
			Name   string `json:"name"`
			Fields []struct {
				Name string `json:"name"`
				Type struct {
					Name   string
					Kind   string
					OfType struct {
						Name string
						Kind string
					} `json:"ofType"`
				} `json:"type"`
			} `json:"fields"`
		} `json:"__type"`
	}
	if err := b.Gql.Run(context.Background(), req, &res); err != nil {
		log.Fatal(err)
	}

	resText, _ := json.MarshalIndent(res, "", "  ")

	msg.Respond(pre([]string{string(resText)}))
	return nil
}
