// Package slap executes shell commands and streams their output to Slack.
package slap

// Source identifies where a line of output originated.
type Source int

const (
	Stdout Source = iota
	Stderr
)

func (s Source) String() string {
	switch s {
	case Stdout:
		return "stdout"
	case Stderr:
		return "stderr"
	default:
		return "unknown"
	}
}

// Line is a single line of captured output tagged with its source.
type Line struct {
	Source Source
	Text   string
}
