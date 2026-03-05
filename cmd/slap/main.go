// The slap command executes a shell command and streams its output to a
// Slack channel via an incoming webhook.
package main

import (
	"fmt"
	"os"

	"github.com/ajbeck/slack-stdout-pipe/internal/slap"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	webhookURL := os.Getenv("SLAP_TARGET")
	if webhookURL == "" {
		fmt.Fprintln(os.Stderr, "slap: SLAP_TARGET environment variable is not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: slap <command> [args...]")
		os.Exit(1)
	}

	name := os.Args[1]
	args := os.Args[2:]

	s := slap.New(webhookURL, Version)
	os.Exit(s.Run(name, args))
}
