package zaptest

import "go.uber.org/zap"

func examples(logger *zap.Logger, sugar *zap.SugaredLogger) {
	logger.Info("server started")  // ok
	logger.Info("Server started")  // want `lower-case`
	logger.Warn("запуск")          // want `non-ASCII`

	sugar.Infof("connected to %s", "db")  // ok
	sugar.Infof("Connected to %s", "db")  // want `lower-case`
	sugar.Infow("request done")           // ok
	sugar.Infow("Request done")           // want `lower-case`
}
