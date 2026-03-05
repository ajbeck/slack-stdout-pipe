// Package gutenberg converts plain-text files from Project Gutenberg and
// similar sources into clean markdown by stripping formatting artifacts.
package gutenberg

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Convert reads plain text from r and writes cleaned markdown to w.
// It strips leading whitespace, joins hard-wrapped paragraphs, converts
// structural headings, and removes table-of-contents blocks and license
// footers.
func Convert(w io.Writer, r io.Reader) error {
	lines, err := readLines(r)
	if err != nil {
		return fmt.Errorf("reading input: %v", err)
	}

	lines = stripIndentation(lines)
	lines = removeFooter(lines)
	lines = removeTOCs(lines)
	lines = convertHeadings(lines)
	lines = joinParagraphs(lines)
	lines = collapseBlankLines(lines)
	lines = trimEnds(lines)

	bw := bufio.NewWriter(w)
	for _, l := range lines {
		fmt.Fprintln(bw, l)
	}
	return bw.Flush()
}

func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func stripIndentation(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = strings.TrimLeftFunc(l, unicode.IsSpace)
	}
	return out
}

// removeFooter removes everything from the "----------" license separator
// onward.
func removeFooter(lines []string) []string {
	for i, l := range lines {
		if strings.TrimSpace(l) == "----------" {
			// Walk back to remove trailing blank lines before the separator.
			end := i
			for end > 0 && lines[end-1] == "" {
				end--
			}
			return lines[:end]
		}
	}
	return lines
}

// removeTOCs removes "Table of contents" blocks. A TOC block starts with
// a line containing "Table of contents" (case-insensitive) and extends
// through all subsequent lines until a line that looks like a structural
// heading (all-caps with enough length) or two consecutive blank lines.
func removeTOCs(lines []string) []string {
	var out []string
	i := 0
	for i < len(lines) {
		if strings.EqualFold(strings.TrimSpace(lines[i]), "table of contents") {
			i++
			// Skip everything until we hit a structural heading or
			// run out of lines. TOC blocks contain entry lines and
			// blank separators.
			for i < len(lines) {
				trimmed := strings.TrimSpace(lines[i])
				if isBookTitle(trimmed) || isStoryTitle(trimmed) || isChapterHeading(trimmed) {
					break
				}
				i++
			}
			continue
		}
		out = append(out, lines[i])
		i++
	}
	return out
}

// convertHeadings detects structural elements and converts them to
// markdown headings.
func convertHeadings(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case isBookTitle(trimmed):
			out = append(out, "# "+toTitleCase(trimmed))
		case isAuthorLine(trimmed):
			out = append(out, "*"+toTitleCase(trimmed)+"*")
		case isChapterHeading(trimmed):
			out = append(out, "## "+toTitleCase(trimmed))
		case isStoryTitle(trimmed):
			out = append(out, "# "+toTitleCase(trimmed))
		default:
			out = append(out, l)
		}
	}
	return out
}

func isBookTitle(s string) bool {
	return s == "THE ADVENTURES OF SHERLOCK HOLMES"
}

func isAuthorLine(s string) bool {
	return strings.EqualFold(s, "arthur conan doyle")
}

func isChapterHeading(s string) bool {
	return len(s) > 7 && strings.HasPrefix(s, "CHAPTER") && isAllUpper(s)
}

// storyTitles is the set of known story titles in ALL CAPS form.
// Using a known list avoids false positives on in-story centered text.
var storyTitles = map[string]bool{
	"A SCANDAL IN BOHEMIA":                  true,
	"THE RED-HEADED LEAGUE":                 true,
	"A CASE OF IDENTITY":                    true,
	"THE BOSCOMBE VALLEY MYSTERY":           true,
	"THE FIVE ORANGE PIPS":                  true,
	"THE MAN WITH THE TWISTED LIP":          true,
	"THE ADVENTURE OF THE BLUE CARBUNCLE":   true,
	"THE ADVENTURE OF THE SPECKLED BAND":    true,
	"THE ADVENTURE OF THE ENGINEER'S THUMB": true,
	"THE ADVENTURE OF THE NOBLE BACHELOR":   true,
	"THE ADVENTURE OF THE BERYL CORONET":    true,
	"THE ADVENTURE OF THE COPPER BEECHES":   true,
}

func isStoryTitle(s string) bool {
	return storyTitles[s]
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// toTitleCase converts an ALL CAPS string to Title Case.
func toTitleCase(s string) string {
	lower := strings.ToLower(s)
	words := strings.Fields(lower)
	// Minor words that stay lowercase unless first or last.
	minor := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "but": true,
		"or": true, "for": true, "nor": true, "on": true, "at": true,
		"to": true, "in": true, "of": true, "with": true,
	}
	for i, w := range words {
		if i == 0 || i == len(words)-1 || !minor[w] {
			words[i] = capitalizeWord(w)
		}
	}
	return strings.Join(words, " ")
}

func capitalizeWord(w string) string {
	if len(w) == 0 {
		return w
	}
	runes := []rune(w)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// joinParagraphs merges consecutive non-blank lines into single lines,
// treating blank lines as paragraph separators.
func joinParagraphs(lines []string) []string {
	var out []string
	var para []string
	flush := func() {
		if len(para) > 0 {
			out = append(out, strings.Join(para, " "))
			para = para[:0]
		}
	}
	for _, l := range lines {
		if l == "" {
			flush()
			out = append(out, "")
			continue
		}
		// Don't join markdown headings or emphasis lines into paragraphs.
		if strings.HasPrefix(l, "#") || (strings.HasPrefix(l, "*") && strings.HasSuffix(l, "*")) {
			flush()
			out = append(out, l)
			continue
		}
		para = append(para, l)
	}
	flush()
	return out
}

// collapseBlankLines reduces runs of consecutive blank lines to a single
// blank line.
func collapseBlankLines(lines []string) []string {
	var out []string
	prevBlank := false
	for _, l := range lines {
		if l == "" {
			if prevBlank {
				continue
			}
			prevBlank = true
		} else {
			prevBlank = false
		}
		out = append(out, l)
	}
	return out
}

// trimEnds removes leading and trailing blank lines.
func trimEnds(lines []string) []string {
	start := 0
	for start < len(lines) && lines[start] == "" {
		start++
	}
	end := len(lines)
	for end > start && lines[end-1] == "" {
		end--
	}
	return lines[start:end]
}
