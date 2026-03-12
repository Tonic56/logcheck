package lowercase

import (
	"log/slog"
	"log"
)

func examples() {
	slog.Info("starting server") // ok
	slog.Info("Starting server") // want `log message must begin with a lower-case letter`

	log.Print("connected to db") // ok
	log.Print("Connected to db") // want `log message must begin with a lower-case letter`

	slog.Error("failed to dial") // ok
	slog.Error("Failed to dial") // want `log message must begin with a lower-case letter`
}
