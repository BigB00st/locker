package main

import (
	"fmt"

	"gitlab.com/amit-yuval/locker/cli/command"
	"gitlab.com/amit-yuval/locker/cmd"
	"gitlab.com/amit-yuval/locker/signal"
	"gitlab.com/amit-yuval/locker/utils"
)

func main() {
	go signal.HandleSignals()
	if utils.IsChild() {
		if err := command.Child(); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		cmd.Execute(cmd.GetCmd())
	}
}
