package nospecial

import "log/slog"

func examples() {
	slog.Info("server started")           // ok
	slog.Info("server started!")          // want `noisy special character`
	slog.Error("connection failed!!!")    // want `noisy special character`
	slog.Warn("something went wrong...")  // want `repeated punctuation`
	slog.Warn("warning: something wrong") // want `noisy special character`
}
