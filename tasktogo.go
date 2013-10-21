package main

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"io"
	"os"
	"os/user"
	"path"
)

var (
	Version = "0.1"

	// DefaultListLocation is the location of the default task list,
	// which is ".tasktogo" in at the user's home directory. Getting
	// that in the var declaration requires a function call to wrap
	// os/user.Current().HomeDir
	DefaultListLocation = path.Join(
		func() string {
			user, err := user.Current()
			if err != nil {
				glog.Warningf("Could not get current user: %s\n", err)
				return "."
			}

			if user.HomeDir == "" {
				glog.Warningf("Could not get user's homedir: %s\n", err)
				return "."
			}

			return user.HomeDir
		}(),
		".tasktogo")

	// Ctx is the global context.
	Ctx *Context
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

	// List is task List that contains all known tasks and associated
	// data.
	// TODO: ensure that it is sorted
	List

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
	// Set up a basic context. In the future, this could be determined
	// by flags.
	Ctx = &Context{
		Input:  bufio.NewReader(os.Stdin),
		Output: os.Stdout,
	}

	// Attempt to load default task list.
	var err error
	Ctx.List, err = ReadListFile(DefaultListLocation)
	if err != nil {
		msg := fmt.Sprintf("Could not read task list: %s\n", err)
		glog.Error(msg)
		writePrompt(Ctx, msg)
	} else {
		// If there were no errors, record information in the context.
		Ctx.loadpath = DefaultListLocation
	}

	// If there are arguments, run in command mode.
	if len(os.Args) > 1 {
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
	c, err := ParseCommand(os.Args[1:])
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
