package goidiomatic

// Sections: accepted expressions; source admission limits; parse errors;
// independently bounded AST traversal.

import (
	"errors"
	"strings"
	"testing"

	"go.starlark.net/syntax"
)

func TestParseSourceAccepted(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"number", "1"},
		{"total_bytes_boundary", "1" + strings.Repeat(" ", maxSourceBytes-1)},
		{"code_bytes_boundary", `"` + strings.Repeat("a", maxSourceCodeBytes-2) + `"`},
		{"unicode_bytes_boundary", `"` + strings.Repeat("é", (maxSourceCodeBytes-2)/2) + `"`},
		{"json_whitespace", " \t\r\n1 \t\r\n"},
		{"unicode_string", `"你好"`},
		{"unicode_identifier", "名字"},
		{"comments", "# comment\n1 # tail\n"},
		{"raw_string", `r"\n"`},
		{"bytes", `b"\x00"`},
		{"triple_string", "\"\"\"first\nsecond\"\"\""},
		{"lambda", "(lambda x: x + 1)(2)"},
		{"comprehension", "[x for x in items if x]"},
		{"compound_data", `({"x": [1, 2]}, items[1:], set([1]))`},
		{"depth_boundary", strings.Repeat("(", maxSourceASTDepth-1) + "1" + strings.Repeat(")", maxSourceASTDepth-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, phase, err := parseSource(tt.source)
			if err != nil || phase != "" || expr == nil {
				t.Fatalf("parseSource = (%T, %q, %v); want expression and no error", expr, phase, err)
			}
		})
	}
}

func TestParseSourceAdmissionLimits(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"total_bytes", "1" + strings.Repeat(" ", maxSourceBytes), "4096 bytes"},
		{"code_bytes", strings.Repeat("a", maxSourceCodeBytes+1), "512 non-whitespace bytes"},
		{"literal_bytes", `"` + strings.Repeat("a", maxSourceCodeBytes-1) + `"`, "512 non-whitespace bytes"},
		{"comment_bytes", "#" + strings.Repeat("a", maxSourceCodeBytes) + "\n1", "512 non-whitespace bytes"},
		{"unicode_counts_bytes", `"` + strings.Repeat("é", maxSourceCodeBytes/2) + `"`, "512 non-whitespace bytes"},
		{"vertical_tab_is_not_exempt", strings.Repeat("\v", maxSourceCodeBytes+1), "512 non-whitespace bytes"},
		{"form_feed_is_not_exempt", strings.Repeat("\f", maxSourceCodeBytes+1), "512 non-whitespace bytes"},
		{"nbsp_is_not_exempt", strings.Repeat("\u00a0", maxSourceCodeBytes), "512 non-whitespace bytes"},
		// These are deliberately rejected before invoking the recursive parser.
		{"not_chain", strings.Repeat("not ", maxSourceCodeBytes/3+1) + "1", "512 non-whitespace bytes"},
		{"minus_chain", strings.Repeat("-", maxSourceCodeBytes+1) + "1", "512 non-whitespace bytes"},
		{"plus_chain", strings.Repeat("+", maxSourceCodeBytes+1) + "1", "512 non-whitespace bytes"},
		{"tilde_chain", strings.Repeat("~", maxSourceCodeBytes+1) + "1", "512 non-whitespace bytes"},
		{"lambda_chain", strings.Repeat("lambda: ", maxSourceCodeBytes/7+1) + "1", "512 non-whitespace bytes"},
		{"ternary_chain", strings.Repeat("1 if 1 else ", maxSourceCodeBytes/8+1) + "0", "512 non-whitespace bytes"},
		{"paren_chain", strings.Repeat("(", maxSourceCodeBytes/2+1) + "1" + strings.Repeat(")", maxSourceCodeBytes/2+1), "512 non-whitespace bytes"},
		{"ast_depth", strings.Repeat("(", maxSourceASTDepth) + "1" + strings.Repeat(")", maxSourceASTDepth), "depth 64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, phase, err := parseSource(tt.source)
			if err == nil || phase != "source" || expr != nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("parseSource = (%T, %q, %v); want source error containing %q", expr, phase, err, tt.want)
			}
		})
	}
}

func TestParseSourceParseErrors(t *testing.T) {
	for _, source := range []string{"", " \t\r\n", "1 +", "x = 1", "1; 2", "'unterminated", "[(1]", "load(\"x\", \"a\")", string([]byte{0xff})} {
		t.Run(source, func(t *testing.T) {
			expr, phase, err := parseSource(source)
			var syntaxErr syntax.Error
			if expr != nil || phase != "parse" || !errors.As(err, &syntaxErr) {
				t.Fatalf("parseSource = (%T, %q, %T %v); want original syntax.Error", expr, phase, err, err)
			}
		})
	}
}

func TestCheckSourceASTNodeLimit(t *testing.T) {
	// Construct a flat tree directly: the admission byte limit is tighter
	// than most source spellings of this defensive AST node boundary.
	for _, count := range []int{maxSourceASTNodes - 1, maxSourceASTNodes, maxSourceASTNodes + 1} {
		expr := &syntax.ListExpr{List: make([]syntax.Expr, count)}
		for i := range expr.List {
			expr.List[i] = &syntax.Ident{Name: "item"}
		}
		err := checkSourceAST(expr)
		if count == maxSourceASTNodes-1 && err != nil {
			t.Fatalf("tree with exactly %d nodes rejected: %v", maxSourceASTNodes, err)
		}
		if count >= maxSourceASTNodes && (err == nil || !strings.Contains(err.Error(), "512 nodes")) {
			t.Fatalf("tree over node bound = %v, want node-limit error", err)
		}
	}
}
