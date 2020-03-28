package main

import (
	"fmt"

	"gitlab.com/amit-yuval/locker/cli/command"
	"gitlab.com/amit-yuval/locker/cmd"
	"gitlab.com/amit-yuval/locker/utils"
)

func main() {
	if utils.IsChild() {
		if err := command.Child(); err != nil {
			fmt.Println(err)
		}
	} else {
		if err := cmd.Execute(cmd.GetCmd()); err != nil {
			fmt.Println(err)
		}
	}
}
