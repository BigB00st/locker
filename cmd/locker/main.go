package main

import (
	"fmt"

	"gitlab.com/amit-yuval/locker/internal/cli/command"

	"gitlab.com/amit-yuval/locker/internal/signal"
	"gitlab.com/amit-yuval/locker/internal/utils"
)

func main() {
	go signal.HandleSignals()
	if utils.IsChild() {
		if err := command.Child(); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		Execute(GetCmd())
	}
}
