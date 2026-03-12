// Package logmatch handles detection and argument extraction for supported
// logging library calls.

package logmatch

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// Call carries the extracted data from a detected logging call.
type Call struct {
	// CallPos is the position of the call expression.
	CallPos token.Pos

	// MsgArg is the AST node for the message argument.
	MsgArg ast.Expr

	// MsgText holds the concatenated string-literal parts of the argument.
	// Variable parts are absent; the value is empty for pure-variable args.
	MsgText string

	// RawExpr is the source text of the entire message argument.
	RawExpr string

	// IsSimpleLiteral is true when MsgArg is a plain *ast.BasicLit STRING.
	IsSimpleLiteral bool
}

// logPackages maps import paths of supported logging packages to true.
var logPackages = map[string]bool{
	"log":             true,
	"log/slog":        true,
	"go.uber.org/zap": true,
}

// logMethods is the set of method names that carry a message as their first
// argument across all supported loggers.
var logMethods = map[string]bool{
	// Common across slog + zap
	"Info": true, "Warn": true, "Error": true, "Debug": true,
	// zap extras
	"Fatal": true, "Panic": true, "DPanic": true,
	// slog context variants
	"InfoContext": true, "WarnContext": true, "ErrorContext": true, "DebugContext": true,
	"InfoCtx": true, "WarnCtx": true, "ErrorCtx": true, "DebugCtx": true,
	// standard log
	"Print": true, "Printf": true, "Println": true,
	"Fatalf": true, "Fatalln": true, "Panicf": true, "Panicln": true,
}

// Extract attempts to recognise expr as a log call from a supported library.
// It returns (Call, true) on success and (Call, false) otherwise.
func Extract(pass interface {
	TypesInfo() *types.Info
}, fset *token.FileSet, expr *ast.CallExpr) (Call, bool) {
	sel, ok := expr.Fun.(*ast.SelectorExpr)
	if !ok {
		return Call{}, false
	}

	if !logMethods[sel.Sel.Name] {
		return Call{}, false
	}

	obj := pass.TypesInfo().Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil {
		return Call{}, false
	}

	pkg := obj.Pkg().Path()
	if !logPackages[pkg] && !strings.HasPrefix(pkg, "go.uber.org/zap") {
		return Call{}, false
	}

	if len(expr.Args) == 0 {
		return Call{}, false
	}

	msgArg := expr.Args[0]
	text, raw := resolveArg(msgArg)

	_, isLit := msgArg.(*ast.BasicLit)

	return Call{
		CallPos:         expr.Pos(),
		MsgArg:          msgArg,
		MsgText:         text,
		RawExpr:         raw,
		IsSimpleLiteral: isLit,
	}, true
}

// resolveArg walks the argument expression tree and returns:
//   - text: concatenated value of any string literal parts
//   - raw: best-effort source representation of the full expression
func resolveArg(expr ast.Expr) (text, raw string) {
	var parts []string
	collectLiterals(expr, &parts)
	text = strings.Join(parts, "")
	raw = exprText(expr)
	return
}

func collectLiterals(expr ast.Expr, out *[]string) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			s := unquote(e.Value)
			*out = append(*out, s)
		}
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			collectLiterals(e.X, out)
			collectLiterals(e.Y, out)
		}
	case *ast.ParenExpr:
		collectLiterals(e.X, out)
	}
}

func exprText(expr ast.Expr) string {
	var b strings.Builder
	ast.Inspect(expr, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			b.WriteString(v.Value)
		case *ast.Ident:
			b.WriteString(v.Name)
		case *ast.SelectorExpr:
			// write X.Sel
		}
		return true
	})
	return b.String()
}

func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	switch s[0] {
	case '"':
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, `\"`, `"`)
		s = strings.ReplaceAll(s, `\\`, `\`)
		s = strings.ReplaceAll(s, `\n`, "\n")
		s = strings.ReplaceAll(s, `\t`, "\t")
	case '`':
		s = s[1 : len(s)-1]
	}
	return s
}
