package rules

import (
	"fmt"
	"strings"
	"unicode"
)

// NoSpecialCharsRule rejects log messages that contain emoji, Unicode
// pictographic symbols, or a configurable set of ASCII punctuation characters
// that add noise without conveying structured information.
type NoSpecialCharsRule struct {
	// Permitted contains any ASCII characters the user explicitly allows,
	// sourced from the logcheck.yaml field `allowed_chars`.
	Permitted string
}

func (NoSpecialCharsRule) Name() string { return "no-special-chars" }


const noisyASCII = "!@#$%^&*+=|\\<>?`~;':\"" // note: colon included

func (r NoSpecialCharsRule) Evaluate(msg LogMessage) *Diagnostic {
	allowed := make(map[rune]struct{}, len(r.Permitted))
	for _, ch := range r.Permitted {
		allowed[ch] = struct{}{}
	}

	for _, ch := range msg.Text {
		if _, ok := allowed[ch]; ok {
			continue
		}

		if isPictographic(ch) {
			return &Diagnostic{
				RuleName: "no-special-chars",
				Message:  fmt.Sprintf("log message contains emoji or pictographic symbol %q", ch),
				Suggestion: buildStripFix(msg),
			}
		}

		if ch <= unicode.MaxASCII && strings.ContainsRune(noisyASCII, ch) {
			return &Diagnostic{
				RuleName: "no-special-chars",
				Message:  fmt.Sprintf("log message contains noisy special character %q", ch),
				Suggestion: buildStripFix(msg),
			}
		}
	}

	// Separately catch repeated punctuation like "!!!" or "..." which isn't
	// caught by the per-character loop above.
	if seq := findRepeatedPunct(msg.Text); seq != "" {
		return &Diagnostic{
			RuleName: "no-special-chars",
			Message:  fmt.Sprintf("log message contains repeated punctuation %q", seq),
		}
	}

	return nil
}

// isPictographic returns true for rune ranges that cover emoji, miscellaneous
// symbols and other pictographic code-points.
func isPictographic(r rune) bool {
	return unicode.Is(unicode.So, r) ||
		(r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F)
}

// findRepeatedPunct returns the first repeated-punctuation sequence found, or "".
func findRepeatedPunct(s string) string {
	for _, seq := range []string{"...", "!!", "??", "***"} {
		if strings.Contains(s, seq) {
			return seq
		}
	}
	return ""
}

// buildStripFix constructs a Fix that removes emoji and noisy characters.
// Only emitted for simple string literals to avoid surprising rewrites.
func buildStripFix(msg LogMessage) *Fix {
	if !msg.SimpleLiteral {
		return nil
	}
	cleaned := stripNoisyChars(msg.Text)
	if cleaned == msg.Text {
		return nil
	}
	return &Fix{
		Summary: "remove emoji and noisy characters",
		NewText: []byte(`"` + cleaned + `"`),
	}
}

// stripNoisyChars removes pictographic runes and noisy ASCII from s.
func stripNoisyChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isPictographic(r) {
			continue
		}
		if r <= unicode.MaxASCII && strings.ContainsRune(noisyASCII, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
