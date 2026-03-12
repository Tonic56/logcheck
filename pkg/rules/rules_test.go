package rules_test

import (
	"testing"

	"github.com/Tonic56/logcheck/pkg/rules"
)

// makeMsg is a helper that creates a LogMessage with only Text and RawExpr set.
func makeMsg(text string) rules.LogMessage {
	return rules.LogMessage{Text: text, RawExpr: text, SimpleLiteral: true}
}

func makeMsgExpr(text, raw string) rules.LogMessage {
	return rules.LogMessage{Text: text, RawExpr: raw, SimpleLiteral: false}
}

// ---------------------------------------------------------------------------
// LowercaseRule
// ---------------------------------------------------------------------------

func TestLowercase_Pass(t *testing.T) {
	r := rules.LowercaseRule{}
	cases := []string{
		"starting server",
		"failed to connect",
		"123 items processed",
		"",
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got != nil {
			t.Errorf("LowercaseRule.Evaluate(%q) = %v; want nil", c, got.Message)
		}
	}
}

func TestLowercase_Fail(t *testing.T) {
	r := rules.LowercaseRule{}
	cases := []string{
		"Starting server",
		"Failed to connect",
		"ERROR occurred",
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got == nil {
			t.Errorf("LowercaseRule.Evaluate(%q) = nil; want violation", c)
		}
	}
}

func TestLowercase_AutoFix(t *testing.T) {
	r := rules.LowercaseRule{}
	msg := rules.LogMessage{Text: "Starting server", SimpleLiteral: true}
	diag := r.Evaluate(msg)
	if diag == nil {
		t.Fatal("expected violation, got nil")
	}
	if diag.Suggestion == nil {
		t.Fatal("expected auto-fix suggestion, got nil")
	}
	want := `"starting server"`
	if string(diag.Suggestion.NewText) != want {
		t.Errorf("fix NewText = %q; want %q", diag.Suggestion.NewText, want)
	}
}

// ---------------------------------------------------------------------------
// EnglishRule
// ---------------------------------------------------------------------------

func TestEnglish_Pass(t *testing.T) {
	r := rules.EnglishRule{}
	cases := []string{
		"connected to database",
		"retry attempt 3 of 5",
		"", // empty always passes
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got != nil {
			t.Errorf("EnglishRule.Evaluate(%q) = %v; want nil", c, got.Message)
		}
	}
}

func TestEnglish_Fail(t *testing.T) {
	r := rules.EnglishRule{}
	cases := []string{
		"запуск сервера",
		"ошибка подключения к базе данных",
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got == nil {
			t.Errorf("EnglishRule.Evaluate(%q) = nil; want violation", c)
		}
	}
}

// ---------------------------------------------------------------------------
// NoSpecialCharsRule
// ---------------------------------------------------------------------------

func TestNoSpecial_Pass(t *testing.T) {
	r := rules.NoSpecialCharsRule{}
	cases := []string{
		"server started",
		"connected to db",
		"retry 3/5",
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got != nil {
			t.Errorf("NoSpecialCharsRule.Evaluate(%q) = %v; want nil", c, got.Message)
		}
	}
}

func TestNoSpecial_Emoji(t *testing.T) {
	r := rules.NoSpecialCharsRule{}
	if got := r.Evaluate(makeMsg("server started🚀")); got == nil {
		t.Error("expected violation for emoji, got nil")
	}
}

func TestNoSpecial_NoisyPunct(t *testing.T) {
	r := rules.NoSpecialCharsRule{}
	cases := []string{"connection failed!", "warning!!!"}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got == nil {
			t.Errorf("NoSpecialCharsRule.Evaluate(%q) = nil; want violation", c)
		}
	}
}

func TestNoSpecial_RepeatedPunct(t *testing.T) {
	r := rules.NoSpecialCharsRule{}
	cases := []string{"something went wrong...", "really???"}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got == nil {
			t.Errorf("NoSpecialCharsRule.Evaluate(%q) = nil; want violation", c)
		}
	}
}

func TestNoSpecial_Permitted(t *testing.T) {
	r := rules.NoSpecialCharsRule{Permitted: "!"}
	// "!" is now allowed.
	if got := r.Evaluate(makeMsg("connection failed!")); got != nil {
		t.Errorf("expected no violation with '!' permitted, got %v", got.Message)
	}
}

// ---------------------------------------------------------------------------
// SensitiveDataRule
// ---------------------------------------------------------------------------

func TestSensitive_Pass(t *testing.T) {
	r := &rules.SensitiveDataRule{Keywords: rules.DefaultSensitiveKeywords()}
	cases := []struct{ text, raw string }{
		{"user authenticated successfully", "\"user authenticated successfully\""},
		{"token validated", "\"token validated\""},
		{"session expired", "\"session expired\""},
		{"api request completed", "\"api request completed\""},
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsgExpr(c.text, c.raw)); got != nil {
			t.Errorf("SensitiveDataRule.Evaluate(%q) = %v; want nil", c.text, got.Message)
		}
	}
}

func TestSensitive_Fail_LiteralKeyword(t *testing.T) {
	r := &rules.SensitiveDataRule{Keywords: rules.DefaultSensitiveKeywords()}
	cases := []string{
		"user password: secret",
		"api_key=value",
	}
	for _, c := range cases {
		if got := r.Evaluate(makeMsg(c)); got == nil {
			t.Errorf("SensitiveDataRule.Evaluate(%q) = nil; want violation", c)
		}
	}
}

func TestSensitive_Fail_VariableName(t *testing.T) {
	r := &rules.SensitiveDataRule{Keywords: rules.DefaultSensitiveKeywords()}
	// The text looks harmless but the raw expression contains a sensitive var.
	msg := makeMsgExpr("request info", `"request info" + jwtToken`)
	if got := r.Evaluate(msg); got == nil {
		t.Error("SensitiveDataRule: expected violation for jwtToken variable name, got nil")
	}
}

func TestSensitive_CustomPattern(t *testing.T) {
	r := &rules.SensitiveDataRule{
		Keywords:       rules.DefaultSensitiveKeywords(),
		CustomPatterns: []string{`sk_live_[a-zA-Z0-9]+`},
	}
	// Should fire on custom pattern match.
	if got := r.Evaluate(makeMsg("stripe key sk_live_abcXYZ123")); got == nil {
		t.Error("SensitiveDataRule: expected violation for custom pattern, got nil")
	}
	// Should not fire when pattern does not match.
	if got := r.Evaluate(makeMsg("stripe key configured")); got != nil {
		t.Errorf("SensitiveDataRule: unexpected violation: %v", got.Message)
	}
}

func TestSensitive_InvalidCustomPatternIgnored(t *testing.T) {
	// An invalid regexp should not panic — it is silently skipped.
	r := &rules.SensitiveDataRule{
		Keywords:       rules.DefaultSensitiveKeywords(),
		CustomPatterns: []string{`[invalid(`},
	}
	if got := r.Evaluate(makeMsg("anything here")); got != nil {
		t.Errorf("invalid custom pattern caused unexpected violation: %v", got.Message)
	}
}

func TestSensitive_ExtraKeyword(t *testing.T) {
	kws := append(rules.DefaultSensitiveKeywords(), "internal_key")
	r := &rules.SensitiveDataRule{Keywords: kws}
	if got := r.Evaluate(makeMsg("using internal_key value")); got == nil {
		t.Error("SensitiveDataRule: expected violation for custom keyword, got nil")
	}
}
