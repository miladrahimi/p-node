package cmd

import (
	"fmt"
	"github.com/miladrahimi/p-node/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "start",
		Run: func(_ *cobra.Command, _ []string) {
			a, err := app.New()
			defer a.Shutdown()
			if err != nil {
				panic(fmt.Sprintf("%+v\n", err))
			}
			a.Init()
			a.Xray.RunWithConfig()
			a.HttpServer.Run()
			a.Wait()
		},
	})
}
