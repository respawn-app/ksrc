package main

import (
	"fmt"
	"os"

	"github.com/respawn-app/ksrc/internal/cli"
)

func main() {
	app := cli.NewApp()
	cmd := cli.NewRootCommand(app)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
