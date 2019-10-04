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
)

func (b *Bot) CmdInvalid(msg joe.Message) error {
	// TODO: enable this after fix is merged https://github.com/go-joe/slack-adapter/pull/6
	//b.Logger.Info("Invalid Command", zap.String("user", msg.AuthorID), zap.String("command", msg.Text))
	//msg.Respond("Hi! I'm a bot and `%s` in an invalid command. Try `help`", msg.Text)
	return nil
}

func (b *Bot) CmdHelp(msg joe.Message) error {
	var resp = []string{
		"Here's how you can interract with me",
		"------------------------------------",
		"hi: Say hi!",
		"help: you're reading it",
		"version: reports the bot version",
		"",
		"get clusters: List clusters",
		"",
		"get users: List users",
		"get user <username>: Show user information",
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

func (b *Bot) CmdGetUsers(msg joe.Message) error {
	req := graphql.NewRequest(`{
		users: users_v1 {
			name
			github_username
			redhat_username
			slack_username
			path
		}
	}`)
	req.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

	var res struct {
		Users []struct {
			Name           string `json:"name"`
			GithubUsername string `json:"github_username"`
			RedhatUsername string `json:"redhat_username"`
			SlackUsername  string `json:"slack_username"`
			Path           string `json:"path"`
		} `json:"users"`
	}
	if err := b.Gql.Run(context.Background(), req, &res); err != nil {
		log.Fatal(err)
	}

	var users []string
	for _, u := range res.Users {
		userText := fmt.Sprintf("%s: %s",
			u.RedhatUsername,
			u.Name,
		)
		users = append(users, userText)
	}
	sort.Strings(users)

	msg.Respond(pre(users))

	return nil
}

func (b *Bot) CmdGetUser(msg joe.Message) error {
	req := graphql.NewRequest(`{
		users: users_v1 {
			name
			github_username
			redhat_username
			slack_username
			path
			public_gpg_key
		}
	}`)
	req.Header.Set("Authorization", "Basic "+b.GqlBasicAuth)

	var res struct {
		Users []struct {
			Name           string `json:"name"`
			GithubUsername string `json:"github_username"`
			RedhatUsername string `json:"redhat_username"`
			SlackUsername  string `json:"slack_username"`
			Path           string `json:"path"`
			PublicGpgKey   string `json:"public_gpg_key"`
		} `json:"users"`
	}
	if err := b.Gql.Run(context.Background(), req, &res); err != nil {
		log.Fatal(err)
	}

	for _, u := range res.Users {
		if u.RedhatUsername == msg.Matches[0] {
			resText := []string{
				u.Path,
				fmt.Sprintf("Name: %s", u.Name),
				fmt.Sprintf("Username: %s", u.RedhatUsername),
			}
			if u.PublicGpgKey != "" {
				resText = append(resText, fmt.Sprintf("GPG pubkey:\n%s", u.PublicGpgKey))

			}
			msg.Respond(pre(resText))
			return nil
		}
	}

	msg.Respond("User %s not found", msg.Matches[0])
	return nil
}

func (b *Bot) CmdGetBotUsers(msg joe.Message) error {
	users, err := b.Auth.GetUsers()
	if err != nil {
		msg.RespondE("There was an error while retrieving users")
		return err
	}
	msg.Respond(pre(users))
	return nil
}

func (b *Bot) CmdGetBotUser(msg joe.Message) error {
	userID := msg.Matches[0]
	perms, err := b.Auth.GetUserPermissions(userID)
	if err != nil {
		msg.RespondE("There was an error while retrieving user permissions: %v", err)
		return err
	}
	resp := []string{
		fmt.Sprintf("User: %s", userID)
		fmt.Sprintf("Permissions: %s", perms)
	}
	msg.Respond(pre(resp))
	return nil
}
