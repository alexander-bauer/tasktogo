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

var (
	ErrNoArguments    = errors.New("no arguments given")
	ErrUnknownCommand = errors.New("unknown command")
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
	"done": (*Command).CmdDone,
	"d":    (*Command).CmdDone,
}

// ParseCommand constructs a command based on a set of arguments,
// including the zeroth, and returns any errors.
func ParseCommand(args []string) (c *Command, err error) {
	// Initialize the Command that will be returned.
	c = &Command{}

	// If there are no arguments, return an error, so that we don't
	// run into a panic later.
	if len(args) == 0 {
		return c, ErrNoArguments
	}

	// Check for the existence of the matching function, and return an
	// error if it's not found.
	var ok bool
	c.Run, ok = RunMap[strings.ToLower(args[0])]
	if !ok {
		return c, ErrUnknownCommand
	}

	c.Args = args[1:]
	return
}

type Runner func(*Command, *Context) error

func (c *Command) CmdHelp(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked help")

	fmt.Fprintf(ctx.Output, "TaskToDo version %s\n\n", Version)
	fmt.Fprintf(ctx.Output, "    help\t\t\t\t\t- print this menu\n")
	fmt.Fprintf(ctx.Output, "    list\t\t\t\t\t- list all tasks\n")
	fmt.Fprintf(ctx.Output, "    add [Task name] [priority] [Month day hour:minute]\t- add a task\n")
	fmt.Fprintf(ctx.Output, "    done [Task name]\t\t\t\t- complete a task\n")

	return nil
}

func (c *Command) CmdList(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked list")

	for _, task := range ctx.List {
		_, err = io.WriteString(ctx.Output, task.String())
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

func (c *Command) CmdDone(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked done")

	// Re-combine the arguments into a single prefix string to search
	// for.
	searchterm := strings.ToLower(strings.Join(c.Args, " "))

	// Iterate through the List and remove the first Task for which
	// the searchterm matches the start of the string.
	for n, task := range ctx.List {
		if strings.HasPrefix(
			strings.ToLower(task.Name), searchterm) {
			// Reslice around the task to be removed.
			ctx.List = append(ctx.List[:n], ctx.List[n+1:]...)
			ctx.modified = true
			return nil
		}
	}
	return nil
}
