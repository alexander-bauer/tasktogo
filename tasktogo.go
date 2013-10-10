package main

import (
	"bufio"
	"github.com/golang/glog"
	"io"
	"os"
)

var (
	Version = "0"
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
}

func main() {
	// Set up a basic context. In the future, this could be determined
	// by flags.
	ctx := &Context{
		Input:  bufio.NewReader(os.Stdin),
		Output: os.Stdout,
	}

	for {
		// Print the prompt once, and get any errors.
		cmd, err := Prompt(ctx)

		if err == io.EOF {
			// If we encounter a graceful EOF, then we must exit
			// immediately.
			glog.V(1).Infoln("Encountered graceful EOF - exiting")

			// Print a final newline before exiting, however, so that
			// the shell prompt isn't affected.
			writePrompt(ctx, "\n")

			// Ensure that log data is written before exiting.
			glog.Flush()
			os.Exit(0)
		} else if err != nil && cmd == nil {
			// If the error was not graceful and cannot be recovered
			// from, exit fatally.
			writePrompt(ctx, "Fatal error: %s\n", err)

			// Ensure that log data is written before exiting.
			glog.Flush()
			glog.Fatalf("Error: %s\n", err)
		} else if err != nil {
			// If the error was user-related, as implied by cmd not
			// being nil, just log and output the error.
			writePrompt(ctx, "Error: %s\n", err)
			glog.Warningf("User error: %s\n", err)
		}

		err = cmd.Run(cmd, ctx)
		if err != nil {
			writePrompt(ctx, "Error: %s\n", err)
			glog.Warningf("Error in command: %s\n", err)
		}
	}

	// Ensure that log data is written.
	glog.Flush()
}
