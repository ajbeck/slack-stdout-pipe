package slap

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Run executes the named command with args, streaming output to both the
// terminal and the Slack webhook at webhookURL. It returns the command's
// exit code.
func Run(webhookURL string, name string, args []string) int {
	lines := make(chan Line, 256)
	sender := NewSender(webhookURL, lines)
	go sender.Run()

	cmdStr := name + " " + strings.Join(args, " ")
	sender.SendMessage(fmt.Sprintf(":rocket: `%s` started", cmdStr))

	start := time.Now()
	exitCode := execute(name, args, lines)
	elapsed := time.Since(start)

	// Channel closed by execute; wait for sender to drain.
	sender.Wait()

	emoji := ":white_check_mark:"
	if exitCode != 0 {
		emoji = ":x:"
	}
	// Send final status as a separate post after the sender has drained
	// all output, so it appears last.
	finalSender := NewSender(webhookURL, nil)
	finalSender.SendMessage(fmt.Sprintf(
		"%s `%s` exited %d in %s",
		emoji, cmdStr, exitCode, elapsed.Round(time.Millisecond),
	))

	return exitCode
}

// execute runs the command, tees output to the terminal, sends lines to
// the channel, and returns the exit code. It closes lines when done.
func execute(name string, args []string, lines chan<- Line) int {
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
	go scanPipe(&wg, stdoutPipe, os.Stdout, Stdout, lines)
	go scanPipe(&wg, stderrPipe, os.Stderr, Stderr, lines)
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
