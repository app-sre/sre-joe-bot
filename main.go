package main

import (
	"log"
	"os"
	"strings"

	"github.com/jfchevrette/sre-joe-bot/bot"
)

func main() {
	bot, err := bot.NewBot("sre-joe-bot",
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

	bot.RespondRegex("^[Hh]i", bot.CmdHi)
	bot.RespondRegex("^[Hh]elp", bot.CmdHelp)
	bot.RespondRegex("^[Cc]luster [Ll]ist", bot.CmdClusters)
	bot.RespondRegex(".*", bot.CmdInvalid)

	err = bot.Run()
	if err != nil {
		bot.Logger.Fatal(err.Error())
	}
}
