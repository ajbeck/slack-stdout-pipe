package slap

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// commander executes a command, sends its output to the lines channel,
// and returns the exit code. It must close lines when done.
type commander interface {
	Run(name string, args []string, lines chan<- Line) int
}

// Slap orchestrates command execution and Slack output streaming.
type Slap struct {
	poster     httpPoster
	commander  commander
	webhookURL string
	version    string
}

// New returns a Slap configured with the real HTTP client and OS command
// execution.
func New(webhookURL, version string) *Slap {
	return &Slap{
		poster:     &http.Client{Timeout: 30 * time.Second},
		commander:  execCommander{stdout: os.Stdout, stderr: os.Stderr},
		webhookURL: webhookURL,
		version:    version,
	}
}

// NewDebug returns a Slap that logs what would be sent to Slack via slog
// instead of making real HTTP requests. Command stdout/stderr is discarded.
func NewDebug(version string) *Slap {
	return &Slap{
		poster:     newLogPoster(),
		commander:  execCommander{stdout: io.Discard, stderr: io.Discard},
		webhookURL: "debug",
		version:    version,
	}
}

// Run executes the named command with args, streaming output to both the
// terminal and the Slack webhook. It returns the command's exit code.
func (s *Slap) Run(name string, args []string) int {
	lines := make(chan Line, 256)
	sender := newSender(s.webhookURL, s.poster, lines)
	go sender.Run()

	cmdStr := name + " " + strings.Join(args, " ")
	sender.SendMessage(fmt.Sprintf(":rocket: `%s` started — slap %s", cmdStr, s.version))

	start := time.Now()
	exitCode := s.commander.Run(name, args, lines)
	elapsed := time.Since(start)

	// Channel closed by commander; wait for sender to drain.
	sender.Wait()

	emoji := ":white_check_mark:"
	if exitCode != 0 {
		emoji = ":x:"
	}
	// Send final status as a separate post after the sender has drained
	// all output, so it appears last.
	finalSender := newSender(s.webhookURL, s.poster, nil)
	finalSender.SendMessage(fmt.Sprintf(
		"%s `%s` exited %d in %s",
		emoji, cmdStr, exitCode, elapsed.Round(time.Millisecond),
	))

	return exitCode
}

// execCommander runs OS commands via exec.Command.
type execCommander struct {
	stdout io.Writer
	stderr io.Writer
}

func (c execCommander) Run(name string, args []string, lines chan<- Line) int {
	defer close(lines)

	cmd := exec.Command(name, args...)
	cmd.Env = filteredEnv()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "slap: stdout pipe: %v\n", err)
		return 1
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "slap: stderr pipe: %v\n", err)
		return 1
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "slap: start: %v\n", err)
		return 1
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scanPipe(&wg, stdoutPipe, c.stdout, Stdout, lines)
	go scanPipe(&wg, stderrPipe, c.stderr, Stderr, lines)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "slap: wait: %v\n", err)
		return 1
	}
	return 0
}

// scanPipe reads from r line-by-line, writes each line to w (terminal),
// and sends a tagged Line on the channel.
func scanPipe(wg *sync.WaitGroup, r io.Reader, w io.Writer, src Source, lines chan<- Line) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Fprintln(w, text)
		lines <- Line{Source: src, Text: text}
	}
}

// filteredEnv returns the current environment with SLAP_TARGET removed.
func filteredEnv() []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, "SLAP_TARGET=") {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}
