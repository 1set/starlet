package internal

// Migrate from https://github.com/makenowjust/heredoc/blob/main/heredoc.go

import (
	"fmt"
	"strings"
)

const maxInt = int(^uint(0) >> 1)

// HereDoc returns un-indented string as here-document.
func HereDoc(raw string) string {
	skipFirstLine := false
	if len(raw) > 0 && raw[0] == '\n' {
		raw = raw[1:]
	} else {
		skipFirstLine = true
	}

	rawLines := strings.Split(raw, "\n")
	lines := rawLines
	if skipFirstLine {
		lines = lines[1:]
	}

	minIndentSize := getMinIndent(lines)
	lines = removeIndentation(lines, minIndentSize)

	return strings.Join(rawLines, "\n")
}

// isSpace checks whether the rune represents space or not.
// Only white spcaes (U+0020) and horizontal tabs are treated as space character.
// It is the same as Go.
//
// See https://github.com/MakeNowJust/heredoc/issues/6#issuecomment-524231625.
func isSpace(r rune) bool {
	switch r {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

// getMinIndent calculates the minimum indentation in lines, excluding empty lines.
func getMinIndent(lines []string) int {
	minIndentSize := maxInt

	for i, line := range lines {
		indentSize := 0
		for _, r := range line {
			if isSpace(r) {
				indentSize++
			} else {
				break
			}
		}

		if len(line) == indentSize {
			if i == len(lines)-1 && indentSize < minIndentSize {
				lines[i] = ""
			}
		} else if indentSize < minIndentSize {
			minIndentSize = indentSize
		}
	}
	return minIndentSize
}

// removeIndentation removes n characters from the front of each line in lines.
func removeIndentation(lines []string, n int) []string {
	for i, line := range lines {
		if len(lines[i]) >= n {
			lines[i] = line[n:]
		}
	}
	return lines
}

// HereDocf returns unindented and formatted string as here-document.
// Formatting is done as for fmt.Printf().
func HereDocf(raw string, args ...interface{}) string {
	return fmt.Sprintf(HereDoc(raw), args...)
}
