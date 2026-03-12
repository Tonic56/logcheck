package sensitive

import "log/slog"

func examples(password, apiKey, jwtToken string) {
	slog.Info("user authenticated successfully")          // ok — safe context word
	slog.Info("token validated")                          // ok — safe context word after keyword
	slog.Info("user password " + password)                // want `may expose sensitive data`
	slog.Debug("api request completed")                   // ok
	slog.Debug("apikey " + apiKey)                        // want `may expose sensitive data`
	slog.Info("sending request with token " + jwtToken)   // want `may expose sensitive data`
}
