package main

import (
	"log"
	"os"
	"strings"

	"github.com/jfchevrette/sre-joe-bot/bot"
)

const (
	botVersion = "0.0.1"
)

func main() {
	bot, err := bot.NewBot("sre-joe-bot",
		bot.WithVersionCommand(botVersion),
		bot.WithSlackAdapter(os.Getenv("SLACK_TOKEN")),
		bot.WithGraphqlClient(
			os.Getenv("APPINTERFACE_URL"),
			os.Getenv("APPINTERFACE_AUTH"),
		),
	)
	if err != nil {
		log.Fatalf("could not create bot: %+v\n", err)
	}

	adminIDS := strings.Split(os.Getenv("BOT_ADMIN_IDS"), ",")
	for _, userID := range adminIDS {
		userID = strings.TrimSpace(userID)
		bot.Auth.Grant("admin", userID)
	}

	bot.Respond("hi", bot.CmdHi)
	bot.Respond("help", bot.CmdHelp)
	bot.Respond("get cluster[s]?", bot.CmdGetClusters)
	bot.Respond("get schema[s]?", bot.CmdGetSchemas)
	bot.Respond("get schema[s]? (.+)", bot.CmdGetSchema)
	bot.Respond("get user[s]?", bot.CmdGetUsers)
	bot.Respond("get user[s]? (.+)", bot.CmdGetUser)

	// TODO: This is not working at the moment. See: https://github.com/go-joe/joe/issues/25
	bot.Respond(".*", bot.CmdInvalid)

	err = bot.Run()
	if err != nil {
		bot.Logger.Fatal(err.Error())
	}
}
