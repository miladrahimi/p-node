package cmd

import (
	"fmt"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"xray-node/internal/app"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: startFunc,
}

func startFunc(_ *cobra.Command, _ []string) {
	if utils.FileExist("./storage/database.json") {
		fmt.Println("here")
		err := os.Rename("./storage/database.json", "./storage/database/app.json")
		fmt.Println(err)
	}
	if utils.FileExist("./storage/xray.json") {
		fmt.Println("here")
		err := os.Rename("./storage/xray.json", "./storage/app/xray.json")
		fmt.Println(err)
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
