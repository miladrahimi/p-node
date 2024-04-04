package cmd

import (
	"fmt"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "p-node",
}

func init() {
	cobra.OnInitialize(func() { fmt.Println(config.AppName) })
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
