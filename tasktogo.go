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
	Version = "0"

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

	for {
		// Print the prompt once, and get any errors.
		cmd, err := Prompt(Ctx)

		if err == io.EOF {
			// If we encounter a graceful EOF, then we must exit
			// immediately.
			glog.V(1).Infoln("Encountered graceful EOF - exiting")

			// Print a final newline before exiting, however, so that
			// the shell prompt isn't affected.
			writePrompt(Ctx, "\n")

			// Exit gracefully.
			exit(0)
		} else if err != nil && cmd == nil {
			// If the error was not graceful and cannot be recovered
			// from, exit fatally.
			writePrompt(Ctx, "Fatal error: %s\n", err)

			// Ensure that log data is written before exiting.
			glog.Flush()
			glog.Fatalf("Error: %s\n", err)
		} else if err != nil {
			// If the error was user-related, as implied by cmd not
			// being nil, log and output the error, and start from the
			// beginning of the loop.
			writePrompt(Ctx, "Error: %s\n", err)
			glog.Warningf("User error: %s\n", err)
			continue
		}

		err = cmd.Run(cmd, Ctx)
		if err != nil {
			writePrompt(Ctx, "Error: %s\n", err)
			glog.Warningf("Error in command: %s\n", err)
		}
	}

	exit(0)
}
