package cmd

import (
	"fmt"

	"github.com/miladrahimi/p-node/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "serve",
		Run: serve,
	})

	// deprecated: use serve instead
	rootCmd.AddCommand(&cobra.Command{
		Use: "start",
		Run: serve,
	})
}

// serve runs the application and xray.
func serve(_ *cobra.Command, _ []string) {
	a, err := app.New()
	defer func() {
		a.Close()
	}()
	if err != nil {
		panic(fmt.Sprintf("%+v\n", err))
	}

	if err = a.Run(); err != nil {
		panic(fmt.Sprintf("%+v\n", err))
	}

	a.Wait()
}
