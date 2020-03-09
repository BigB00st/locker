package main

import (
	"gitlab.com/amit-yuval/locker/cli/command"
	"gitlab.com/amit-yuval/locker/cmd"
	"gitlab.com/amit-yuval/locker/utils"
)

func main() {
	if utils.IsChild() {
		command.Child()
	} else {
		cmd.Execute(cmd.GetCmd())
	}
}
