package main

import (
	"fmt"
	"github.com/golang/glog"
)

import ()

type Command struct {
	// Run is the function underlying the Command, and can be called
	// to execute the behavior of the Command.
	Run Runner

	// Args is a list of the remaining arguments (not including the
	// primary command), separated by spaces, obeying quotes, double
	// quotes, and backslashes.
	Args []string
}

// RunMap is used to map command strings to Runners.
var RunMap = map[string]Runner{
	"help": (*Command).CmdHelp,
	"h":    (*Command).CmdHelp,
}

type Runner func(*Command, *Context) error

func (c *Command) CmdHelp(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked help")

	fmt.Fprintf(ctx.Output, "TaskToDo version %s\n\n", Version)
	fmt.Fprintf(ctx.Output, "    help\t\t\t- print this menu\n")

	return nil
}
