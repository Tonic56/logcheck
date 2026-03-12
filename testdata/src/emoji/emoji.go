package emoji

import "log/slog"

func examples() {
	slog.Info("server started")    // ok
	slog.Info("server started🚀")  // want `emoji or pictographic symbol`
}
