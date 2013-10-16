package main

import (
	"fmt"
	"github.com/golang/glog"
	"io"
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
	"list": (*Command).CmdList,
	"l":    (*Command).CmdList,
}

type Runner func(*Command, *Context) error

func (c *Command) CmdHelp(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked help")

	fmt.Fprintf(ctx.Output, "TaskToDo version %s\n\n", Version)
	fmt.Fprintf(ctx.Output, "    help\t\t\t- print this menu\n")
	fmt.Fprintf(ctx.Output, "    list\t\t\t- list all tasks\n")

	return nil
}

func (c *Command) CmdList(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked list")

	for _, task := range ctx.List {
		_, err = io.WriteString(ctx.Output, fmt.Sprintf(
			"%s - %s (%d)\n\t%s\n",
			task.Name, task.DueBy.Format("Monday, Jan 02, 15:04"),
			task.Priority, task.Description))
		if err != nil {
			glog.Warningf("Error listing tasks: %s\n", err)
		}
	}

	return nil
}
