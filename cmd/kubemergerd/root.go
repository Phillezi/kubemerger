package main

import (
	"fmt"
	"log"
	"os"

	viperconf "github.com/Phillezi/common/config/viper"
	"github.com/phillezi/kubemerger/internal/daemon"
	"github.com/phillezi/kubemerger/internal/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:     short,
	Long:    long,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stderr, "%s\nVersion: %s\n", long, version)
		d := daemon.New(daemon.WithContext(cmd.Context()), daemon.WithRoot(viper.GetString("root-dir")), daemon.WithOutput(viper.GetString("output")))
		if err := d.Run(); err != nil {
			log.Default().Printf("Error: %v\n", err)
		}
	},
}

func init() {
	cobra.OnInitialize(func() { viperconf.InitConfig("kubemerger") })

	rootCmd.Flags().String("root-dir", defaults.DefaultKubeDir, "The directory to watch recursively")
	_ = viper.BindPFlag("root-dir", rootCmd.Flags().Lookup("root-dir"))
	rootCmd.Flags().String("output", defaults.DefaultKubeConfig, "The output kubeconfig path")
	_ = viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
}
