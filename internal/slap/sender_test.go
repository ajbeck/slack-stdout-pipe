package slap

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestUnitFormatBatch(t *testing.T) {
	tests := []struct {
		name     string
		batch    []Line
		wantN    int
		contains []string
	}{
		{
			name:  "empty",
			batch: nil,
			wantN: 0,
		},
		{
			name:     "single_line",
			batch:    []Line{{Source: Stdout, Text: "hello"}},
			wantN:    1,
			contains: []string{"```", "[stdout] hello"},
		},
		{
			name: "mixed_sources",
			batch: []Line{
				{Source: Stdout, Text: "out"},
				{Source: Stderr, Text: "err"},
			},
			wantN:    1,
			contains: []string{"[stdout] out", "[stderr] err"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBatch(tt.batch)
			if len(got) != tt.wantN {
				t.Errorf("formatBatch(%v) returned %d messages, want %d", tt.batch, len(got), tt.wantN)
			}
			combined := strings.Join(got, "")
			for _, s := range tt.contains {
				if !strings.Contains(combined, s) {
					t.Errorf("formatBatch() output = %q, want to contain %q", combined, s)
				}
			}
		})
	}
}

func TestUnitFormatBatchSplit(t *testing.T) {
	bigLine := strings.Repeat("x", maxMessageLen)
	batch := []Line{
		{Source: Stdout, Text: bigLine},
		{Source: Stdout, Text: "second"},
	}

	got := formatBatch(batch)
	if len(got) < 2 {
		t.Errorf("formatBatch() = %d messages, want at least 2 for oversized batch", len(got))
	}

	// Each message must be a complete code block.
	for i, msg := range got {
		if !strings.HasPrefix(msg, "```\n") {
			t.Errorf("message[%d] does not start with opening code fence: %q", i, msg[:20])
		}
		if !strings.HasSuffix(msg, "```") {
			t.Errorf("message[%d] does not end with closing code fence", i)
		}
	}
}

func TestUnitBackoffDuration(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		want       time.Duration
	}{
		{
			name:       "valid_header",
			retryAfter: "3",
			want:       3 * time.Second,
		},
		{
			name: "missing_header",
			want: defaultBackoff,
		},
		{
			name:       "non_numeric",
			retryAfter: "abc",
			want:       defaultBackoff,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{Header: make(http.Header)}
			if tt.retryAfter != "" {
				resp.Header.Set("Retry-After", tt.retryAfter)
			}
			got := backoffDuration(resp)
			if got != tt.want {
				t.Errorf("backoffDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
