// Package rules defines the Rule interface and all built-in rule implementations.

package rules

import "go/token"

// LogMessage is passed to every Rule.Check call and contains all the
// information needed to evaluate a log argument.
type LogMessage struct {
	// Text is the resolved string content (literal parts joined).
	Text string

	// RawExpr is the source text of the whole argument expression,
	// e.g. `"token: " + jwtToken`. Rules that scan for sensitive variable
	// names should use this field.
	RawExpr string

	// Pos / End mark the argument node so auto-fixes can replace it.
	Pos token.Pos
	End token.Pos

	// SimpleLiteral is true when the argument is a plain *ast.BasicLit string,
	// meaning we can safely auto-fix it without touching surrounding code.
	SimpleLiteral bool
}

// Fix describes an automatic source-level correction.
type Fix struct {
	// Summary is shown in golangci-lint / IDE output.
	Summary string
	// NewText replaces the bytes from Pos to End.
	NewText []byte
	// Pos and End from the original LogMessage should be used by the caller.
}

// Diagnostic is the finding returned when a rule is violated.
// A nil Diagnostic means the message is acceptable.
type Diagnostic struct {
	// RuleName is the canonical name of the rule that fired.
	RuleName string
	// Message is the human-readable explanation.
	Message string
	// Suggestion, when non-nil, provides an auto-correction.
	Suggestion *Fix
}

// Rule is the interface every logcheck rule must implement.
type Rule interface {
	// Name returns the kebab-case identifier used in logcheck.yaml.
	Name() string

	// Evaluate checks the given log message and returns a Diagnostic if the
	// message violates this rule, or nil if it is acceptable.
	Evaluate(msg LogMessage) *Diagnostic
}
