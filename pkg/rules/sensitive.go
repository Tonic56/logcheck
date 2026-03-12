package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// SensitiveDataRule checks log messages for patterns that suggest the presence
// of credentials, tokens, or other sensitive values.

type SensitiveDataRule struct {
	// Keywords is the list of case-insensitive sensitive terms.
	Keywords []string

	// CustomPatterns is a list of user-defined Go regular expressions.
	// Any match in the resolved message text triggers a violation.
	CustomPatterns []string

	// compiled holds the per-keyword word-boundary regexp, built lazily.
	compiled []*regexp.Regexp
	// compiledCustom holds compiled CustomPatterns regexps.
	compiledCustom []*regexp.Regexp
}

func (r *SensitiveDataRule) Name() string { return "no-sensitive-data" }

// DefaultSensitiveKeywords returns the built-in keyword list.
func DefaultSensitiveKeywords() []string {
	return []string{
		"password", "passwd", "secret", "token",
		"apikey", "api_key", "authkey", "auth_key",
		"credential", "private_key", "access_key",
		"session", "jwt", "bearer", "ssn", "credit_card",
	}
}

// safeContextWords are word tokens that, when appearing immediately after a
// sensitive keyword, indicate the message is about the credential's status
// rather than its value. This reduces common false positives like
// "token validated" or "session expired".
var safeContextWords = map[string]bool{
	"validated": true, "expired":    true, "refreshed":  true,
	"revoked":   true, "rotated":    true, "created":    true,
	"deleted":   true, "updated":    true, "generated":  true,
	"ok":        true, "success":    true, "successful": true,
	"failed":    true, "invalid":    true, "missing":    true,
	"present":   true, "enabled":    true, "disabled":   true,
	"error":     true, "authorized": true, "unauthorized": true,
}

func (r *SensitiveDataRule) ensureCompiled() {
	if len(r.compiled) == len(r.Keywords) && len(r.compiledCustom) == len(r.CustomPatterns) {
		return
	}
	r.compiled = make([]*regexp.Regexp, len(r.Keywords))
	for i, kw := range r.Keywords {
		pattern := `(?i)\b` + regexp.QuoteMeta(strings.ToLower(kw)) + `\b`
		r.compiled[i] = regexp.MustCompile(pattern)
	}
	r.compiledCustom = make([]*regexp.Regexp, 0, len(r.CustomPatterns))
	for _, pat := range r.CustomPatterns {
		re, err := regexp.Compile(pat)
		if err == nil {
			r.compiledCustom = append(r.compiledCustom, re)
		}
	}
}

func (r *SensitiveDataRule) Evaluate(msg LogMessage) *Diagnostic {
	r.ensureCompiled()

	lowerText := strings.ToLower(msg.Text)
	words := strings.Fields(lowerText)

	for i, re := range r.compiled {
		kw := strings.ToLower(r.Keywords[i])

		// --- Pass 1: whole-word match in the resolved message text ---
		if re.MatchString(lowerText) {
			// Check if the next word makes this a safe status message.
			idx := re.FindStringIndex(lowerText)
			if idx != nil {
				after := strings.Fields(lowerText[idx[1]:])
				if len(after) > 0 && safeContextWords[after[0]] {
					goto checkExpr
				}
			}
			_ = words
			return &Diagnostic{
				RuleName: "no-sensitive-data",
				Message:  fmt.Sprintf("log message may expose sensitive data (keyword %q detected)", kw),
			}
		}

	checkExpr:
		// --- Pass 2: identifier scan in the raw source expression ---
		// Strip string literal contents so only variable/field names remain,
		// then normalise camelCase and snake_case before matching.
		identifiers := extractIdentifiers(msg.RawExpr)
		normKw := normalize(kw)
		for _, id := range identifiers {
			if strings.Contains(normalize(id), normKw) {
				return &Diagnostic{
					RuleName: "no-sensitive-data",
					Message:  fmt.Sprintf("log message may expose sensitive data (identifier %q looks like %q)", id, kw),
				}
			}
		}
	}

	// --- Pass 3: user-defined custom regexp patterns ---
	for _, re := range r.compiledCustom {
		if re.MatchString(msg.Text) {
			return &Diagnostic{
				RuleName: "no-sensitive-data",
				Message:  fmt.Sprintf("log message matched custom sensitive-data pattern %q", re.String()),
			}
		}
	}

	return nil
}

// normalize folds an identifier to lower-case and removes underscores so that
// camelCase and snake_case variants are equivalent:
//
//	"jwtToken" → "jwttoken"
//	"jwt_token" → "jwttoken"
func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

// extractIdentifiers strips string-literal contents from a Go expression and
// returns the remaining word tokens (variable names, field names, etc.).
func extractIdentifiers(expr string) []string {
	// Remove double-quoted string content.
	inStr := false
	esc := false
	var b strings.Builder
	for _, r := range expr {
		if esc {
			esc = false
			continue
		}
		if r == '\\' && inStr {
			esc = true
			continue
		}
		if r == '"' {
			inStr = !inStr
			b.WriteRune(' ')
			continue
		}
		if !inStr {
			b.WriteRune(r)
		}
	}
	cleaned := b.String()

	// Split on anything that is not a letter, digit or underscore.
	return strings.FieldsFunc(cleaned, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') &&
			!(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') &&
			r != '_'
	})
}
