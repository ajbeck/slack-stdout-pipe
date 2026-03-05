package slap

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// logPoster satisfies httpPoster by logging each payload via slog instead
// of making real HTTP requests.
type logPoster struct {
	logger *slog.Logger
}

func newLogPoster() *logPoster {
	return &logPoster{
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

func (p *logPoster) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %v", err)
	}

	var sp slackPayload
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("unmarshalling payload: %v", err)
	}

	p.logger.Info("slack message", "text", sp.Text)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

const asciiArt = `
      ___  _          ___ _
     / __|| |  __ _  | _ \ |
     \__ \| |_/ _' | |  _/_|
     |___/|____\__,_| |_| (_)

`

// PrintVersion writes version information and ASCII art to stdout.
func PrintVersion(version string) {
	fmt.Print(asciiArt)
	fmt.Printf("  slap %s\n\n", version)
}
