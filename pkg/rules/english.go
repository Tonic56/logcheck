package rules

import (
	"fmt"
	"unicode"
)

// EnglishRule rejects log messages that contain characters outside the ASCII
// range (excluding a small allow-list of typographic symbols). Non-ASCII text
// can break log parsers and aggregators that expect UTF-8 encoding to be
// limited to ASCII in log messages.
type EnglishRule struct{}

func (EnglishRule) Name() string { return "english-only" }

// typographicExtras is a small set of non-ASCII code-points commonly found in
// English technical writing that we intentionally permit.
var typographicExtras = map[rune]bool{
	'\u2013': true, // en-dash
	'\u2014': true, // em-dash
	'\u2026': true, // horizontal ellipsis
	'\u2018': true, // left single quotation
	'\u2019': true, // right single quotation
	'\u201C': true, // left double quotation
	'\u201D': true, // right double quotation
}

func (EnglishRule) Evaluate(msg LogMessage) *Diagnostic {
	for _, r := range msg.Text {
		if r <= unicode.MaxASCII || typographicExtras[r] {
			continue
		}
		return &Diagnostic{
			RuleName: "english-only",
			Message:  fmt.Sprintf("log message contains non-ASCII character %q (only English text is allowed)", r),
		}
	}
	return nil
}
