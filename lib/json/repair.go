package json

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/1set/starlet/internal/jsonrepair"
	"go.starlark.net/starlark"
)

// fenceRe matches a fenced code block, capturing its inner content. LLM
// output commonly wraps JSON in ```json ... ``` fences.
var fenceRe = regexp.MustCompile("(?s)```(?:json|JSON)?[ \\t]*\\r?\\n?(.*?)```")

// generateRepair builds the json.repair / json.try_repair builtins. repair
// recovers a valid JSON *text* from the messy output models produce — code
// fences, surrounding prose, single quotes, trailing commas, comments,
// Python True/None, truncation — for the caller to then json.decode. It
// returns text (not a value): json.decode(json.repair(x)) is the idiom.
//
// Already-valid JSON is returned byte-for-byte unchanged (idempotent), and
// only genuinely-broken input reaches the repair engine — the vendored
// jsonrepair can otherwise double-escape some valid escape sequences, so it
// is never run on text that already parses.
func generateRepair(try bool) func(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var text string
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "text", &text); err != nil {
			if try {
				return starlark.Tuple{none, starlark.String(err.Error())}, nil
			}
			return none, err
		}
		out, err := repairText(text)
		if err != nil {
			if try {
				return starlark.Tuple{none, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("%s: %w", fn.Name(), err)
		}
		if try {
			return starlark.Tuple{starlark.String(out), none}, nil
		}
		return starlark.String(out), nil
	}
}

// repairText normalizes messy text to valid JSON text.
func repairText(text string) (string, error) {
	// Fast path: already-valid JSON is returned unchanged. This makes
	// repair idempotent on good input and avoids running jsonrepair on it
	// (the engine is not perfectly idempotent — e.g. it can double a
	// backslash in some valid escapes).
	if json.Valid([]byte(strings.TrimSpace(text))) {
		return text, nil
	}

	// Level 1: prefer fenced content when a code fence is present.
	candidate := text
	if m := fenceRe.FindStringSubmatch(text); m != nil {
		candidate = m[1]
	}

	// Level 2: take the first balanced {...} or [...] span (drops prose
	// around the JSON; the repair engine errors on "preamble + fenced JSON").
	if span, ok := firstBalancedSpan(candidate); ok {
		candidate = span
	}

	// If the extracted candidate already parses, return it as-is rather
	// than risk the engine mangling valid escapes.
	if json.Valid([]byte(strings.TrimSpace(candidate))) {
		return candidate, nil
	}

	// Level 3: repair (vendored jsonrepair v0.2.2; its go1.19 window is
	// frozen, so the copy is pinned and behavior is golden-locked).
	return jsonrepair.JSONRepair(candidate)
}

// firstBalancedSpan returns the first balanced {...} or [...] substring,
// honoring quoted strings and escapes. A truncated (never-closed) span is
// returned to its end so the repair stage can complete it.
func firstBalancedSpan(s string) (string, bool) {
	rs := []rune(s)
	start, clos := openBracket(rs)
	if start < 0 {
		return "", false
	}
	open := rs[start]
	depth, inStr, esc := 0, false, false
	for i := start; i < len(rs); i++ {
		switch r := rs[i]; {
		case esc:
			esc = false // this char is escaped; consume it
		case inStr && r == '\\':
			esc = true
		case r == '"':
			inStr = !inStr
		case inStr:
			// inside a string literal: brackets are not structural
		case r == open:
			depth++
		case r == clos:
			if depth--; depth == 0 {
				return string(rs[start : i+1]), true
			}
		}
	}
	return string(rs[start:]), true // truncated; repair completes it
}

// openBracket returns the index of the first { or [ in rs and its matching
// closing rune, or (-1, 0) if neither appears.
func openBracket(rs []rune) (int, rune) {
	for i, r := range rs {
		switch r {
		case '{':
			return i, '}'
		case '[':
			return i, ']'
		}
	}
	return -1, 0
}
