package cmd

import (
	"fmt"
	"github.com/miladrahimi/p-manager/pkg/utils"
	"github.com/miladrahimi/p-node/internal/app"
	"github.com/spf13/cobra"
	"os"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: startFunc,
}

func startFunc(_ *cobra.Command, _ []string) {
	if utils.FileExist("./storage/database.json") {
		_ = os.Rename("./storage/database.json", "./storage/database/app.json")
	}
	if utils.FileExist("./storage/xray.json") {
		_ = os.Rename("./storage/xray.json", "./storage/app/xray.json")
	}

	a, err := app.New()
	defer a.Shutdown()
	if err != nil {
		panic(fmt.Sprintf("%+v\n", err))
	}
	a.Init()
	a.Xray.RunWithConfig()
	a.HttpServer.Run()
	a.Wait()
}
