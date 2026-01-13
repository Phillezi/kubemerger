package main

import (
	"context"
	"os"

	"github.com/Phillezi/interrupt/pkg/interrupt"
	"github.com/Phillezi/interrupt/pkg/manager"
)

func main() {
	interrupt.Main(func(m manager.ManagedManager, cancel context.CancelFunc) {
		if err := rootCmd.ExecuteContext(m.Context()); err != nil {
			cancel()
			os.Exit(1)
		}
	}, interrupt.WithManagerOpts(manager.WithPrompt(true, os.Stderr)))
}
