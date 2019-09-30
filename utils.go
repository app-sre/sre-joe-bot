package main

import "strings"

func pre(msgs []string) string {
	return "```\n" + strings.Join(msgs, "\n") + "\n```"
}
