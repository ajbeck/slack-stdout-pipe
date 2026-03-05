// The demo command outputs lines from The Adventures of Sherlock Holmes
// to stdout and stderr at random intervals, for use as sample input to
// slap.
package main

import (
	"embed"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

//go:embed sherlock.md
var sherlockFS embed.FS

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: demo <duration>")
		fmt.Fprintln(os.Stderr, "  example: demo 30s")
		os.Exit(1)
	}

	dur, err := time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "demo: invalid duration: %v\n", err)
		os.Exit(1)
	}

	lines, err := loadLines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "demo: %v\n", err)
		os.Exit(1)
	}

	deadline := time.After(dur)
	i := 0
	for {
		select {
		case <-deadline:
			return
		default:
		}

		line := lines[i%len(lines)]
		i++

		// Randomly pick stdout or stderr.
		if rand.IntN(2) == 0 {
			fmt.Fprintln(os.Stderr, line)
		} else {
			fmt.Fprintln(os.Stdout, line)
		}

		// Random delay between 10ms and 500ms.
		delay := time.Duration(10+rand.IntN(490)) * time.Millisecond
		select {
		case <-deadline:
			return
		case <-time.After(delay):
		}
	}
}

// loadLines reads the embedded sherlock.md and returns non-blank lines.
func loadLines() ([]string, error) {
	data, err := sherlockFS.ReadFile("sherlock.md")
	if err != nil {
		return nil, fmt.Errorf("reading embedded text: %v", err)
	}
	var lines []string
	for l := range strings.SplitSeq(string(data), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no lines found in embedded text")
	}
	return lines, nil
}
