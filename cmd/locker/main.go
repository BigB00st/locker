package main

import (
	"fmt"

	"locker/internal/cli/command"

	"locker/internal/signal"
	"locker/internal/utils"
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
