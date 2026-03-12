package custompat

import "log/slog"

func examples() {
	slog.Info("stripe key configured")              // ok
	slog.Info("stripe key sk_live_abcXYZ123")       // want `matched custom sensitive-data pattern`
	slog.Info("using key gh_pat_ABCdef123456")      // want `matched custom sensitive-data pattern`
}
