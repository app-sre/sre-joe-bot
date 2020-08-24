package main

import (
	"log"
	"os"
	"strings"

	"github.com/app-sre/sre-joe-bot/bot"
	"go.uber.org/zap"
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
		_, err := bot.Auth.Grant("bot.admin", userID)
		if err != nil {
			bot.Logger.Fatal("could not add grant", zap.Error(err))
		}
	}

	bot.Respond("hi", bot.Log(bot.CmdHi))
	bot.Respond("help", bot.Log(bot.CmdHelp))
	bot.Respond("get cluster[s]?", bot.Log(bot.CmdGetClusters))
	bot.Respond("get user[s]?", bot.Log(bot.CmdGetUsers))
	bot.Respond("get user[s]? (.+)", bot.Log(bot.CmdGetUser))

	bot.Respond("get bot user[s]?", bot.Log(bot.Authenticate("bot.admin.read", bot.CmdGetBotUsers)))
	bot.Respond("get bot user[s]? (.+)", bot.Log(bot.Authenticate("bot.admin.read", bot.CmdGetBotUser)))

	bot.Respond("get schema[s]?", bot.Log(bot.Authenticate("bot.admin.read", bot.CmdGetSchemas)))
	bot.Respond("get schema[s]? (.+)", bot.Log(bot.Authenticate("bot.admin.read", bot.CmdGetSchema)))

	// TODO: This is not working at the moment. See: https://github.com/go-joe/joe/issues/25
	bot.Respond(".*", bot.CmdInvalid)

	err = bot.Run()
	if err != nil {
		bot.Logger.Fatal("failed to start bot", zap.Error(err))
	}
}
