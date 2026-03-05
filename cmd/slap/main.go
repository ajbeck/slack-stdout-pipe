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
	debug := os.Getenv("SLAP_BLOCK") == "bingo"

	if debug {
		if len(os.Args) < 2 {
			slap.PrintVersion(Version)
			return
		}
		s := slap.NewDebug(Version)
		os.Exit(s.Run(os.Args[1], os.Args[2:]))
	}

	webhookURL := os.Getenv("SLAP_TARGET")
	if webhookURL == "" {
		fmt.Fprintln(os.Stderr, "slap: SLAP_TARGET environment variable is not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: slap <command> [args...]")
		os.Exit(1)
	}

	s := slap.New(webhookURL, Version)
	os.Exit(s.Run(os.Args[1], os.Args[2:]))
}
