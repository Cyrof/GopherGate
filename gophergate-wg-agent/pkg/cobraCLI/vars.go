package cobraCLI

import "go.uber.org/zap"

var (
	name   string
	peerID string
	Log    *zap.SugaredLogger
)

func SetLogger(l *zap.SugaredLogger) {
	Log = l
}
