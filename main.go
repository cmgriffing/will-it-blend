package main

import (
	"github.com/cmgriffing/will-it-blend/cmd"
)

func main() {
	cmd.Init()
	cmd.RootCmd.Execute()
}
