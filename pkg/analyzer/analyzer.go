// Package analyzer implements the go/analysis pass for logcheck.
//
// It wires together the logmatch detector, the rule set from pkg/rules, and
// the configuration from pkg/config into a single *analysis.Analyzer that can
// be used standalone or as a golangci-lint plugin.
package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/Tonic56/logcheck/pkg/config"
	"github.com/Tonic56/logcheck/pkg/rules"
)

const analyzerName = "logcheck"
const analyzerDoc = "logcheck enforces style, language, punctuation and sensitive-data rules on log messages"

// New builds an *analysis.Analyzer from the given Config.
// Pass nil to use Defaults().
func New(cfg *config.Config) *analysis.Analyzer {
	if cfg == nil {
		cfg = config.Defaults()
	}
	rs := buildRuleSet(cfg)
	return &analysis.Analyzer{
		Name:     analyzerName,
		Doc:      analyzerDoc,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, cfg, rs)
		},
	}
}

// Analyzer is the default instance loaded by the golangci-lint plugin.
// It reads logcheck.yaml from the current working directory.
var Analyzer = New(mustLoadConfig("logcheck.yaml"))

func mustLoadConfig(path string) *config.Config {
	cfg, err := config.Load(path)
	if err != nil {
		return config.Defaults()
	}
	return cfg
}


func buildRuleSet(cfg *config.Config) []rules.Rule {
	sensitiveKws := append(rules.DefaultSensitiveKeywords(), cfg.ExtraKeywords...)

	all := []rules.Rule{
		rules.LowercaseRule{},
		rules.EnglishRule{},
		rules.NoSpecialCharsRule{Permitted: cfg.AllowedChars},
		&rules.SensitiveDataRule{
			Keywords:       sensitiveKws,
			CustomPatterns: cfg.CustomPatterns,
		},
	}

	var active []rules.Rule
	for _, r := range all {
		if cfg.IsEnabled(r.Name()) {
			active = append(active, r)
		}
	}
	return active
}

// ---------------------------------------------------------------------------
// Pass implementation
// ---------------------------------------------------------------------------

// typeInfoAdapter wraps *analysis.Pass so logmatch.Extract can call TypesInfo
// without importing go/analysis directly.
type typeInfoAdapter struct{ pass *analysis.Pass }

func (a typeInfoAdapter) TypesInfo() *types.Info { return a.pass.TypesInfo }

func run(pass *analysis.Pass, _ *config.Config, rs []rules.Rule) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		lc, ok := extractCall(pass, call)
		if !ok {
			return
		}
		applyRules(pass, rs, lc)
	})

	return nil, nil
}

// logCall mirrors logmatch.Call but without the package dependency cycle.
type logCall struct {
	callPos       token.Pos
	msgArg        ast.Expr
	msgText       string
	rawExpr       string
	simpleLiteral bool
}

func extractCall(pass *analysis.Pass, call *ast.CallExpr) (logCall, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return logCall{}, false
	}

	if !isSupportedMethod(sel.Sel.Name) {
		return logCall{}, false
	}

	obj, ok := pass.TypesInfo.Uses[sel.Sel]
	if !ok || obj == nil || obj.Pkg() == nil {
		return logCall{}, false
	}

	if !isSupportedPkg(obj.Pkg().Path()) {
		return logCall{}, false
	}

	if len(call.Args) == 0 {
		return logCall{}, false
	}

	msgArg := call.Args[0]
	text := collectText(msgArg)
	raw := rawSource(pass, msgArg)
	_, isLit := msgArg.(*ast.BasicLit)

	return logCall{
		callPos:       call.Pos(),
		msgArg:        msgArg,
		msgText:       text,
		rawExpr:       raw,
		simpleLiteral: isLit,
	}, true
}

var supportedMethods = map[string]bool{
	// slog + zap shared
	"Info": true, "Warn": true, "Error": true, "Debug": true,
	// zap non-sugar extras
	"Fatal": true, "Panic": true, "DPanic": true,
	// zap SugaredLogger — *f (format) and *w (with fields) variants
	"Infof": true, "Warnf": true, "Errorf": true, "Debugf": true,
	"Fatalf": true, "Panicf": true, "DPanicf": true,
	"Infow": true, "Warnw": true, "Errorw": true, "Debugw": true,
	"Fatalw": true, "Panicw": true, "DPanicw": true,
	// slog context variants
	"InfoContext": true, "WarnContext": true,
	"ErrorContext": true, "DebugContext": true,
	"InfoCtx": true, "WarnCtx": true,
	"ErrorCtx": true, "DebugCtx": true,
	// stdlib log
	"Print": true, "Printf": true, "Println": true,
	"Fatalln": true, "Panicln": true,
}

func isSupportedMethod(name string) bool { return supportedMethods[name] }

func isSupportedPkg(path string) bool {
	switch path {
	case "log", "log/slog":
		return true
	}
	if len(path) >= len("go.uber.org/zap") && path[:len("go.uber.org/zap")] == "go.uber.org/zap" {
		return true
	}
	return false
}

func collectText(expr ast.Expr) string {
	var parts []string
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch v := e.(type) {
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				parts = append(parts, unquoteLit(v.Value))
			}
		case *ast.BinaryExpr:
			if v.Op == token.ADD {
				walk(v.X)
				walk(v.Y)
			}
		case *ast.ParenExpr:
			walk(v.X)
		}
	}
	walk(expr)
	s := ""
	for _, p := range parts {
		s += p
	}
	return s
}

func rawSource(pass *analysis.Pass, expr ast.Expr) string {
	// Walk and reconstruct from AST tokens.
	var b []byte
	ast.Inspect(expr, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			b = append(b, v.Value...)
		case *ast.Ident:
			b = append(b, v.Name...)
		}
		return true
	})
	_ = pass
	return string(b)
}

func unquoteLit(s string) string {
	if len(s) < 2 {
		return s
	}
	switch s[0] {
	case '"':
		s = s[1 : len(s)-1]
		s = replaceEscapes(s)
	case '`':
		s = s[1 : len(s)-1]
	}
	return s
}

func replaceEscapes(s string) string {
	out := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				out = append(out, '\n')
			case 't':
				out = append(out, '\t')
			case '"':
				out = append(out, '"')
			case '\\':
				out = append(out, '\\')
			default:
				out = append(out, s[i], s[i+1])
			}
			i += 2
			continue
		}
		out = append(out, s[i])
		i++
	}
	return string(out)
}

// applyRules runs every active rule against the log call and reports findings.
func applyRules(pass *analysis.Pass, rs []rules.Rule, lc logCall) {
	msg := rules.LogMessage{
		Text:          lc.msgText,
		RawExpr:       lc.rawExpr,
		Pos:           lc.msgArg.Pos(),
		End:           lc.msgArg.End(),
		SimpleLiteral: lc.simpleLiteral,
	}

	for _, r := range rs {
		diag := r.Evaluate(msg)
		if diag == nil {
			continue
		}

		d := analysis.Diagnostic{
			Pos:     lc.msgArg.Pos(),
			End:     lc.msgArg.End(),
			Message: diag.Message,
		}

		if diag.Suggestion != nil {
			d.SuggestedFixes = []analysis.SuggestedFix{
				{
					Message: diag.Suggestion.Summary,
					TextEdits: []analysis.TextEdit{
						{
							Pos:     lc.msgArg.Pos(),
							End:     lc.msgArg.End(),
							NewText: diag.Suggestion.NewText,
						},
					},
				},
			}
		}

		pass.Report(d)
	}
}
