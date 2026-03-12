package analyzer_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/Tonic56/logcheck/pkg/analyzer"
	"github.com/Tonic56/logcheck/pkg/config"
)

// testdataDir returns the absolute path to the testdata directory.
// analysistest treats this as a GOPATH root and looks for packages under
// {testdataDir}/src/{pkgname}, so our testdata layout must be:
//
//	testdata/src/lowercase/
//	testdata/src/english/
//	...
func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata")
}

func allRulesCfg() *config.Config {
	return config.Defaults()
}

func TestAnalyzer_Lowercase(t *testing.T) {
	a := analyzer.New(allRulesCfg())
	analysistest.Run(t, testdataDir(), a, "lowercase")
}

func TestAnalyzer_English(t *testing.T) {
	a := analyzer.New(allRulesCfg())
	analysistest.Run(t, testdataDir(), a, "english")
}

func TestAnalyzer_NoSpecial(t *testing.T) {
	a := analyzer.New(allRulesCfg())
	analysistest.Run(t, testdataDir(), a, "nospecial")
}

// TestAnalyzer_Emoji runs with English rule disabled so the emoji violation
// is caught by no-special-chars rather than english-only (both would fire otherwise).
func TestAnalyzer_Emoji(t *testing.T) {
	cfg := config.Defaults()
	cfg.SetDisabled([]string{"english-only"})
	a := analyzer.New(cfg)
	analysistest.Run(t, testdataDir(), a, "emoji")
}

func TestAnalyzer_Sensitive(t *testing.T) {
	// Disable no-special-chars so only the sensitive rule fires on these messages.
	cfg := config.Defaults()
	cfg.SetDisabled([]string{"no-special-chars"})
	a := analyzer.New(cfg)
	analysistest.Run(t, testdataDir(), a, "sensitive")
}

func TestAnalyzer_DisabledRule(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDisabled([]string{"english-only", "no-special-chars", "no-sensitive-data"})
	a := analyzer.New(cfg)
	analysistest.Run(t, testdataDir(), a, "lowercase")
}

// TestAnalyzer_Zap verifies that zap logger calls (including SugaredLogger *f/*w
// methods) are detected and analysed correctly.
func TestAnalyzer_Zap(t *testing.T) {
	// Only check lowercase + english so the testdata is focused.
	cfg := config.Defaults()
	cfg.SetDisabled([]string{"no-special-chars", "no-sensitive-data"})
	a := analyzer.New(cfg)
	analysistest.Run(t, testdataDir(), a, "zaptest")
}

// TestAnalyzer_CustomPatterns verifies that user-supplied regexp patterns
// in CustomPatterns are applied by the no-sensitive-data rule.
func TestAnalyzer_CustomPatterns(t *testing.T) {
	cfg := config.Defaults()
	cfg.CustomPatterns = []string{`sk_live_[a-zA-Z0-9]+`, `gh_pat_[a-zA-Z0-9]+`}
	// Disable no-special-chars so only the sensitive rule fires.
	cfg.SetDisabled([]string{"no-special-chars"})
	a := analyzer.New(cfg)
	analysistest.Run(t, testdataDir(), a, "custompat")
}
