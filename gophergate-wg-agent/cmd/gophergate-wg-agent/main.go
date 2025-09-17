package main

import (
	stdlog "log"

	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/pkg/cobraCLI"
)

var Version = "0.1.0"

func main() {
	p := paths.ForApp("gophergate-wg-agent")
	if err := p.Ensure(); err != nil {
		stdlog.Fatalf("failed to ensure paths: %v", err)
	}

	logger, flush := logx.Init(logx.Default(paths.AppAgent))

	defer flush()

	logger.Infow("agent started", "version", Version)

	cobraCLI.SetLogger(logger)

	cobraCLI.Execute()
}
