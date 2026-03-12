package rules

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// LowercaseRule checks that log messages start with a lower-case letter.
// Uppercase openings are inconsistent with log-level prefixes added by
// loggers (INFO, ERROR) and make line-by-line scanning harder.
type LowercaseRule struct{}

func (LowercaseRule) Name() string { return "lowercase" }

func (LowercaseRule) Evaluate(msg LogMessage) *Diagnostic {
	s := strings.TrimSpace(msg.Text)
	if s == "" {
		return nil
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || !unicode.IsUpper(r) {
		return nil
	}

	d := &Diagnostic{
		RuleName: "lowercase",
		Message:  "log message must begin with a lower-case letter",
	}

	// Only offer an auto-fix for plain literals to avoid mangling expressions.
	if msg.SimpleLiteral {
		fixed := string(unicode.ToLower(r)) + s[size:]
		d.Suggestion = &Fix{
			Summary: "lowercase the first letter",
			NewText: []byte(`"` + fixed + `"`),
		}
	}

	return d
}
