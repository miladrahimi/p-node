package main

import (
	"fmt"
	"github.com/miladrahimi/p-node/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
