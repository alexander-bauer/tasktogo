package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io"
	"strconv"
	"strings"
	"time"
)

import ()

const (
	DueFormat = "Jan _2 15:04"
)

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
	"add":  (*Command).CmdAdd,
	"a":    (*Command).CmdAdd,
}

type Runner func(*Command, *Context) error

func (c *Command) CmdHelp(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked help")

	fmt.Fprintf(ctx.Output, "TaskToDo version %s\n\n", Version)
	fmt.Fprintf(ctx.Output, "    help\t\t\t\t\t- print this menu\n")
	fmt.Fprintf(ctx.Output, "    list\t\t\t\t\t- list all tasks\n")
	fmt.Fprintf(ctx.Output, "    add [Task name] [priority] [Month day hour:minute]\t- add a task\n")

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

func (c *Command) CmdAdd(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked add")

	t := &Task{}
	var datestring string
	// Separate the arguments into sections and fill out the Task with
	// them. The syntax is "add [multiword name] [priority] [month-day
	// hour:minute]".
	for _, arg := range c.Args {
		// If the Priority has not yet been filled out, try to parse
		// the current argument as an int. Otherwise, append the
		// argument to the datestring to be parsed at the end.
		if t.Priority == 0 {
			t.Priority, err = strconv.Atoi(arg)
			if err != nil {
				// If the priority couldn't be parsed, consider it
				// part of the name.
				t.Name += arg + " "
			}
		} else {
			datestring += arg + " "
		}
	}
	t.Name = strings.TrimRight(t.Name, " ")

	due, err := time.ParseInLocation(DueFormat,
		strings.TrimRight(datestring, " "), time.Local)
	if err != nil {
		return errors.New("Could not parse arguments")
	}
	// Assume that the date given is for this year.
	// TODO: use the next occurence of that date, rather than the one
	// for the current calendar year.
	due = due.AddDate(time.Now().Year(), 0, 0)

	t.DueBy = due.Local()

	// TODO: retrieve a description somehow

	// Now, add the task to the list, sort it, and set the "modified"
	// flag.
	ctx.List = append(ctx.List, t)
	ctx.Sort()
	ctx.modified = true
	return nil
}
