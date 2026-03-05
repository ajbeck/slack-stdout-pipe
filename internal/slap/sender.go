package slap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	coalesceWindow = 500 * time.Millisecond
	// Slack truncates text around 40k chars. Stay well under.
	maxMessageLen  = 38000
	defaultBackoff = 5 * time.Second
)

// httpPoster sends an HTTP POST request and returns the response.
// *http.Client satisfies this interface.
type httpPoster interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// slackPayload is the JSON body sent to a Slack incoming webhook.
type slackPayload struct {
	Text string `json:"text"`
}

// Sender reads Lines from a channel, batches them on a coalescing window,
// and POSTs formatted messages to a Slack webhook.
type Sender struct {
	webhookURL string
	poster     httpPoster
	lines      <-chan Line
	done       chan struct{}
}

// newSender creates a Sender that reads from lines and posts to webhookURL.
func newSender(webhookURL string, poster httpPoster, lines <-chan Line) *Sender {
	return &Sender{
		webhookURL: webhookURL,
		poster:     poster,
		lines:      lines,
		done:       make(chan struct{}),
	}
}

// Run starts the send loop. It returns after the lines channel is closed
// and all buffered output has been sent. Call from its own goroutine.
func (s *Sender) Run() {
	defer close(s.done)

	var batch []Line
	timer := time.NewTimer(coalesceWindow)
	timer.Stop()
	timerRunning := false

	for {
		select {
		case line, ok := <-s.lines:
			if !ok {
				// Channel closed — send remaining batch.
				if len(batch) > 0 {
					s.sendBatch(batch)
				}
				return
			}
			batch = append(batch, line)
			if !timerRunning {
				timer.Reset(coalesceWindow)
				timerRunning = true
			}

		case <-timer.C:
			timerRunning = false
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// Wait blocks until the sender has finished processing all lines.
func (s *Sender) Wait() {
	<-s.done
}

// SendMessage posts a single plain-text message to the webhook.
func (s *Sender) SendMessage(text string) {
	s.post(text)
}

// sendBatch formats a slice of Lines into one or more Slack messages
// and posts them.
func (s *Sender) sendBatch(batch []Line) {
	for _, chunk := range formatBatch(batch) {
		s.post(chunk)
	}
}

// formatBatch converts Lines into code-block-formatted messages,
// splitting if the result would exceed maxMessageLen.
func formatBatch(batch []Line) []string {
	var messages []string
	var buf strings.Builder
	buf.WriteString("```\n")

	for _, l := range batch {
		line := fmt.Sprintf("[%s] %s\n", l.Source, l.Text)
		// If adding this line would exceed the limit, close the block
		// and start a new message.
		if buf.Len()+len(line)+len("```") > maxMessageLen {
			buf.WriteString("```")
			messages = append(messages, buf.String())
			buf.Reset()
			buf.WriteString("```\n")
		}
		buf.WriteString(line)
	}

	if buf.Len() > len("```\n") {
		buf.WriteString("```")
		messages = append(messages, buf.String())
	}
	return messages
}

// post sends a single text message to the Slack webhook, handling
// rate-limit backoff.
func (s *Sender) post(text string) {
	payload, err := json.Marshal(slackPayload{Text: text})
	if err != nil {
		log.Printf("slap: failed to marshal payload: %v", err)
		return
	}

	for {
		resp, err := s.poster.Post(s.webhookURL, "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Printf("slap: webhook POST failed: %v", err)
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := backoffDuration(resp)
			log.Printf("slap: rate limited, retrying in %v", wait)
			time.Sleep(wait)
			continue
		}

		log.Printf("slap: webhook returned %d", resp.StatusCode)
		return
	}
}

// backoffDuration reads Retry-After from the response or returns a default.
func backoffDuration(resp *http.Response) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultBackoff
}
