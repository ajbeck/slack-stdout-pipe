// The gentext command converts a plain-text file to markdown using the
// gutenberg package.
package main

import (
	"fmt"
	"os"

	"github.com/ajbeck/slack-stdout-pipe/gutenberg"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: gentext <input.txt> <output.md>")
		os.Exit(1)
	}

	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "gentext: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "gentext: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	if err := gutenberg.Convert(out, in); err != nil {
		fmt.Fprintf(os.Stderr, "gentext: %v\n", err)
		os.Exit(1)
	}
}
