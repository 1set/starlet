package goidiomatic

import (
	"fmt"

	"go.starlark.net/syntax"
)

const (
	maxSourceBytes     = 4096
	maxSourceCodeBytes = 512
	maxSourceASTDepth  = 64
	maxSourceASTNodes  = 512
)

// parseSource admits a small source before invoking the pinned recursive
// parser, then checks the resulting tree before resolution and compilation.
// Counting literal and comment bytes too is deliberately conservative: large
// data belongs in vars. These limits are not a repair of the upstream parser
// or a hard process memory/CPU boundary.
func parseSource(source string) (syntax.Expr, string, error) {
	if len(source) > maxSourceBytes {
		return nil, "source", fmt.Errorf("source exceeds %d bytes", maxSourceBytes)
	}
	codeBytes := 0
	for i := 0; i < len(source); i++ {
		switch source[i] {
		case ' ', '\t', '\r', '\n':
		default:
			codeBytes++
		}
	}
	if codeBytes > maxSourceCodeBytes {
		return nil, "source", fmt.Errorf("source exceeds %d non-whitespace bytes", maxSourceCodeBytes)
	}

	opts := syntax.FileOptions{Set: true}
	expr, err := opts.ParseExpr("<eval>", source, 0)
	if err != nil {
		return nil, "parse", err
	}
	if err := checkSourceAST(expr); err != nil {
		return nil, "source", err
	}
	return expr, "", nil
}

// checkSourceAST prunes traversal before descending past its own depth bound.
func checkSourceAST(expr syntax.Expr) error {
	var guardErr error
	depth, nodes := 0, 0
	syntax.Walk(expr, func(node syntax.Node) bool {
		if node == nil {
			depth--
			return false
		}
		if guardErr != nil {
			return false
		}
		nodes++
		if nodes > maxSourceASTNodes {
			guardErr = fmt.Errorf("source AST exceeds %d nodes", maxSourceASTNodes)
			return false
		}
		if depth == maxSourceASTDepth {
			guardErr = fmt.Errorf("source AST exceeds depth %d", maxSourceASTDepth)
			return false
		}
		depth++
		return true
	})
	return guardErr
}
