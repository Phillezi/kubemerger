package main

import (
	"log"

	"github.com/phillezi/kubemerger/internal/daemon"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     short,
	Long:    long,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		d := daemon.New(daemon.WithContext(cmd.Context()))
		if err := d.Run(); err != nil {
			log.Default().Printf("Error: %v\n", err)
		}
	},
}

func init() {
	cobra.OnInitialize(func() {})
}
