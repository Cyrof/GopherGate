package main

import (
	"log"

	// "github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/pkg/cobraCLI"
)

func main() {
	p := paths.ForApp("gophergate-wg-agent")
	if err := p.Ensure(); err != nil {
		log.Fatalf("failed to ensure paths: %v", err)
	}

	// logger := logx.Init(logx.Default("gophergate-wg-agent"))

	cobraCLI.Execute()
}
