package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io"
	"os"
	"path"
)

var (
	Version = "0.2.2"

	// Ctx is the global context.
	Ctx *Context
)

// Flags
var (
	FlagColor = flag.Bool("color", true, "enable list colorization")

	FlagList = flag.String("l", path.Join("$HOME", ".tasktogo"),
		"select task list")
)

type Context struct {
	// Input is the bufio.Reader used to drive input. Usually, it will
	// wrap os.Stdin, but could be anything that implements the
	// interface.
	Input *bufio.Reader

	// Output is the io.Writer used to drive output. It will also
	// contain the prompt unless another io.Writer is specified as the
	// Prompt output.
	Output, Prompt io.Writer

	// List is task list that contains all known tasks and associated
	// data.
	List

	// Colors is a flag which determines whether tasks should colorize
	// themselves according to due date when using String().
	Colors bool

	// loadpath is the path on the filesystem from which the List was
	// loaded.
	loadpath string

	// modified is a flag which implies that the List should be saved
	// to its file before exiting.
	modified bool
}

func (ctx *Context) Save() {
	// Only attempt to save if the List has been modified.
	if ctx.modified {
		err := ctx.List.WriteFile(ctx.loadpath)
		if err != nil {
			glog.Errorf("Could not save list: %s\n", err)
		} else {
			glog.V(1).Infof("List saved to %q\n", ctx.loadpath)
		}
	}
}

// exit performs cleanup tasks and exits with the given status code.
func exit(status int) {
	// Save the global context if necessary.
	Ctx.Save()

	glog.Flush()
	os.Exit(status)
}

func main() {
	// Parse and command line flags.
	flag.Parse()

	// Set up a basic context, making use of the flags.
	Ctx = &Context{
		Input:  bufio.NewReader(os.Stdin),
		Output: os.Stdout,

		Colors: *FlagColor,
	}

	// Attempt to load the given task list.
	var err error
	Ctx.loadpath = os.ExpandEnv(*FlagList)
	Ctx.List, err = ReadListFile(Ctx.loadpath)
	if err != nil {
		msg := fmt.Sprintf("Could not read task list: %s\n", err)
		glog.Error(msg)
		writePrompt(Ctx, msg)
	}

	// If there are arguments, run in command mode.
	if flag.NArg() > 0 {
		exit(runCommandMode(Ctx))
	} else {
		exit(runInteractiveMode(Ctx))
	}
}

// runCommandMode constructs a Command from OS arguments, runs it, and
// returns the appropriate exit value.
func runCommandMode(ctx *Context) int {
	// Pass the slice of arguments to TaskToGo to ParseCommand. Note
	// that if there are none, we will pass an empty slice properly,
	// rather than panicing. `([]int{0, 1, 2}[3:]` works perfectly
	// fine.
	c, err := ParseCommand(flag.Args())
	if err != nil {
		// If the command was invalid, log it and return 1.
		writePrompt(ctx, "Error: %s\n", err)
		glog.Warningf("User error: %s\n", err)
		return 1
	}

	// Run the command.
	err = c.Run(c, ctx)
	if err != nil {
		writePrompt(ctx, "Error: %s\n", err)
		glog.Warningf("Error in command: %s\n", err)
		return 1
	}

	// If we get to this point, return successful.
	return 0
}

// runInteractiveMode runs TaskToGo in interactive mode, accepting
// commands from the Context.Input and calling Run on them. It returns
// the value with which the program should exit.
func runInteractiveMode(ctx *Context) int {
	for {
		// Print the prompt once, and get any errors.
		c, err := Prompt(ctx)

		if err == io.EOF {
			// If we encounter a graceful EOF, then we must exit
			// immediately.
			glog.V(1).Infoln("Encountered graceful EOF - exiting")

			// Print a final newline before exiting, however, so that
			// the shell prompt isn't affected.
			writePrompt(ctx, "\n")

			// Exit gracefully.
			exit(0)
		} else if err != nil && c == nil {
			// If the error was not graceful and cannot be recovered
			// from, exit fatally.
			writePrompt(ctx, "Fatal error: %s\n", err)

			// Ensure that log data is written before exiting.
			glog.Flush()
			glog.Fatalf("Error: %s\n", err)
		} else if err != nil {
			// If the error was user-related, as implied by c not
			// being nil, log and output the error, and start from the
			// beginning of the loop.
			writePrompt(ctx, "Error: %s\n", err)
			glog.Warningf("User error: %s\n", err)
			continue
		}

		err = c.Run(c, ctx)
		if err != nil {
			writePrompt(ctx, "Error: %s\n", err)
			glog.Warningf("Error in command: %s\n", err)
		}
	}
}
