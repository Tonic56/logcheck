// Package plugin exposes logcheck as a golangci-lint custom plugin.
//
// Build with:
//
//	go build -buildmode=plugin -o logcheck.so ./plugin
//
// Then reference it in .golangci.yml:
//
//	linters-settings:
//	  custom:
//	    logcheck:
//	      path: logcheck.so
//	      description: Lints log messages for style, language, and sensitive data
//	      original-url: github.com/Tonic56/logcheck
package main

import (
	"github.com/Tonic56/logcheck/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

// AnalyzerPlugin is the golangci-lint plugin entrypoint.
var AnalyzerPlugin analyzerPlugin

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{analyzer.Analyzer}
}
