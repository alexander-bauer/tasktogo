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

	ErrMissingPriority = errors.New("no priority argument given")
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
	"help":       (*Command).CmdHelp,
	"h":          (*Command).CmdHelp,
	"exit":       (*Command).CmdExit,
	"quit":       (*Command).CmdExit,
	"q":          (*Command).CmdExit,
	"list":       (*Command).CmdList,
	"l":          (*Command).CmdList,
	"add":        (*Command).CmdAdd,
	"a":          (*Command).CmdAdd,
	"eventually": (*Command).CmdEventually,
	"e":          (*Command).CmdEventually,
	"done":       (*Command).CmdDone,
	"d":          (*Command).CmdDone,
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
	fmt.Fprintf(ctx.Output, "    exit\t\t\t\t\t- exit gracefully\n")
	fmt.Fprintf(ctx.Output, "    list [maxItems]\t\t\t\t- list all tasks\n")
	fmt.Fprintf(ctx.Output, "    add name priority month day hr:min\t- add a task\n")
	fmt.Fprintf(ctx.Output, "    eventually name priority\t\t- add an eventual task\n")
	fmt.Fprintf(ctx.Output, "    done name\t\t\t\t- complete a task\n")

	return nil
}

func (c *Command) CmdExit(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked exit")
	exit(0)
	return nil
}

func (c *Command) CmdList(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked list")

	// Only show the first n tasks, but make sure that n doesn't go
	// out of bounds. Also, if n is -1, show all tasks.
	var n int

	// If an argument is given, then try to use it.
	if len(c.Args) > 0 {
		n, _ = strconv.Atoi(c.Args[0])
	}

	// If not, then use the context's setting.
	if n == 0 {
		n = ctx.MaxListItems
	}

	if n >= len(ctx.List) || n < 0 {
		n = len(ctx.List) - 1
	}

	for _, task := range ctx.List[:n] {
		_, err = io.WriteString(ctx.Output, task.String())
		if err != nil {
			glog.Warningf("Error listing tasks: %s\n", err)
		}
	}

	return nil
}

func (c *Command) CmdAdd(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked add")

	t := &DefiniteTask{}
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
	// Use the next occurrence of this date by first parsing it as if
	// it's in the current calendar year, or if that is before the
	// current time, then shifting it to the next calendar year.
	currentTime := time.Now()
	due = due.AddDate(currentTime.Year(), 0, 0)
	if due.Before(currentTime) {
		due = due.AddDate(1, 0, 0)
	}

	t.DueBy = due.Local()

	// TODO: retrieve a description somehow

	// Now, add the task to the list, sort it, and set the "modified"
	// flag.
	ctx.List = append(ctx.List, t)
	ctx.Sort()
	ctx.modified = true
	return nil
}

func (c *Command) CmdEventually(ctx *Context) (err error) {
	glog.V(2).Infoln("User invoked eventually")

	t := &EventualTask{}
	// Loop through the arguments until we find a priority factor,
	// which will be just an integer. The syntax is as follows.
	//
	//     eventually [Name] [priority]
	for _, arg := range c.Args {
		// If the Priority has not yet been filled out, try to parse
		// the current argument as an int.
		if t.Priority == 0 {
			t.Priority, err = strconv.Atoi(arg)
			if err != nil {
				// If the priority couldn't be parsed, consider it
				// part of the name.
				t.Name += arg + " "
			}
		}
	}

	// If the priority is still 0, then the syntax was incorrect.
	if t.Priority == 0 {
		return ErrMissingPriority
	}

	t.Name = strings.TrimRight(t.Name, " ")

	// TODO: retrieve a description somehow

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
		if task.Match(searchterm) {
			// Reslice around the task to be removed.
			ctx.List = append(ctx.List[:n], ctx.List[n+1:]...)
			ctx.modified = true
			return nil
		}
	}
	return nil
}
