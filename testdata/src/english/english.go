package english

import "log/slog"

func examples() {
	slog.Info("starting server")       // ok
	slog.Info("запуск сервера")        // want `log message contains non-ASCII character`
	slog.Error("connection failed")    // ok
	slog.Error("ошибка подключения")   // want `log message contains non-ASCII character`
}
