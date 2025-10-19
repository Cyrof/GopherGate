package cobraCLI

import "go.uber.org/zap"

var (
	Log *zap.SugaredLogger
)

func SetLogger(l *zap.SugaredLogger) {
	Log = l
}
