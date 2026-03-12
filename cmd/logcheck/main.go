// Command logcheck is a standalone runner for the logcheck analyzer.
//
// Usage:
//
//	logcheck ./...
//	logcheck -config path/to/logcheck.yaml ./...
package main

import (
	"flag"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Tonic56/logcheck/pkg/analyzer"
	"github.com/Tonic56/logcheck/pkg/config"
)

var configPath = flag.String("config", "logcheck.yaml", "path to logcheck.yaml config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		// Non-fatal: fall back to defaults.
		cfg = config.Defaults()
	}

	singlechecker.Main(analyzer.New(cfg))
}
