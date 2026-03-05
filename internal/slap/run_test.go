package slap

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// stubPoster records all posted payloads and returns a configurable status.
type stubPoster struct {
	mu       sync.Mutex
	payloads []string
	status   int
}

func (p *stubPoster) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	data, _ := io.ReadAll(body)
	var sp slackPayload
	json.Unmarshal(data, &sp)

	p.mu.Lock()
	p.payloads = append(p.payloads, sp.Text)
	p.mu.Unlock()

	status := p.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

func (p *stubPoster) messages() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := make([]string, len(p.payloads))
	copy(cp, p.payloads)
	return cp
}

// stubCommander sends predetermined lines and returns a fixed exit code.
type stubCommander struct {
	lines    []Line
	exitCode int
}

func (c stubCommander) Run(name string, args []string, lines chan<- Line) int {
	defer close(lines)
	for _, l := range c.lines {
		lines <- l
	}
	return c.exitCode
}

func TestUnitRunSuccess(t *testing.T) {
	poster := &stubPoster{}
	s := &Slap{
		poster:     poster,
		commander:  stubCommander{exitCode: 0},
		webhookURL: "https://hooks.example.com/test",
		version:    "1.2.3",
	}

	code := s.Run("make", []string{"build"})
	if code != 0 {
		t.Errorf("Run(make, build) = %d, want 0", code)
	}

	msgs := poster.messages()
	if len(msgs) < 2 {
		t.Fatalf("Run() posted %d messages, want at least 2", len(msgs))
	}

	first := msgs[0]
	if !strings.Contains(first, ":rocket:") {
		t.Errorf("start message = %q, want to contain %q", first, ":rocket:")
	}
	if !strings.Contains(first, "make build") {
		t.Errorf("start message = %q, want to contain %q", first, "make build")
	}
	if !strings.Contains(first, "1.2.3") {
		t.Errorf("start message = %q, want to contain %q", first, "1.2.3")
	}

	last := msgs[len(msgs)-1]
	if !strings.Contains(last, ":white_check_mark:") {
		t.Errorf("final message = %q, want to contain %q", last, ":white_check_mark:")
	}
	if !strings.Contains(last, "exited 0") {
		t.Errorf("final message = %q, want to contain %q", last, "exited 0")
	}
}

func TestUnitRunFailure(t *testing.T) {
	poster := &stubPoster{}
	s := &Slap{
		poster:     poster,
		commander:  stubCommander{exitCode: 42},
		webhookURL: "https://hooks.example.com/test",
		version:    "1.0.0",
	}

	code := s.Run("fail", nil)
	if code != 42 {
		t.Errorf("Run(fail) = %d, want 42", code)
	}

	msgs := poster.messages()
	last := msgs[len(msgs)-1]
	if !strings.Contains(last, ":x:") {
		t.Errorf("final message = %q, want to contain %q", last, ":x:")
	}
	if !strings.Contains(last, "exited 42") {
		t.Errorf("final message = %q, want to contain %q", last, "exited 42")
	}
}

func TestUnitRunCommandOutput(t *testing.T) {
	poster := &stubPoster{}
	s := &Slap{
		poster: poster,
		commander: stubCommander{
			lines: []Line{
				{Source: Stdout, Text: "hello"},
				{Source: Stderr, Text: "world"},
			},
		},
		webhookURL: "https://hooks.example.com/test",
		version:    "1.0.0",
	}

	s.Run("echo", []string{"test"})

	msgs := poster.messages()
	if len(msgs) < 3 {
		t.Fatalf("Run() posted %d messages, want at least 3 (start + output + final)", len(msgs))
	}

	// Batch messages are between the first (start) and last (final).
	var batch string
	for _, m := range msgs[1 : len(msgs)-1] {
		batch += m
	}
	if !strings.Contains(batch, "[stdout] hello") {
		t.Errorf("batch output = %q, want to contain %q", batch, "[stdout] hello")
	}
	if !strings.Contains(batch, "[stderr] world") {
		t.Errorf("batch output = %q, want to contain %q", batch, "[stderr] world")
	}
}
