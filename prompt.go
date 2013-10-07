package main

import (
	"errors"
	"fmt"
	"github.com/gobs/args"
	"strings"
)

const (
	PromptString = ": "
)

var (
	PrimaryMap = map[string]int{
		"help": 1,
	}
)

var (
	ErrNoArguments    = errors.New("no arguments given")
	ErrUnknownPrimary = errors.New("unknown command")
)

// Prompt writes a prompt to the screen and reads the Input stream
// until a linebreak or EOF, then parses that into a Command type. If
// it encounters an error, no the command that's been parsed so far
// will be returned. It will not print any more than the initial
// prompt.
func Prompt(ctx *Context) (cmd *Command, err error) {
	// Write the prompt to the appropriate io.Writer.
	writePrompt(ctx, PromptString)

	// Wait for a line of input. If there is an error, cmd will be
	// nil, indicating that it is an actual error, as opposed to one
	// relating to the command itself.
	line, err := ctx.Input.ReadString('\n')
	if err != nil {
		return
	}

	// Split the arguments, obeying quotes, double quotes, and
	// backslashes.
	inArgs := args.GetArgs(line)

	// If there are no arguments, return an error, so that we don't
	// run into a panic later.
	if len(inArgs) == 0 {
		return nil, ErrNoArguments
	}

	// Initialize the Command that will be returned.
	cmd = &Command{}

	// Check for the existence of the primary command, and return an
	// error if it's not found.
	var ok bool
	cmd.Primary, ok = PrimaryMap[strings.ToLower(inArgs[0])]
	if !ok {
		return cmd, ErrUnknownPrimary
	}

	cmd.Args = inArgs[1:]

	return
}

// writePrompt is a helper function that writes to ctx.Prompt if
// defined, or ctx.Output if not.
func writePrompt(ctx *Context, format string, a ...interface{}) {
	if ctx.Prompt != nil {
		fmt.Fprintf(ctx.Prompt, format, a...)
	} else {
		fmt.Fprintf(ctx.Output, format, a...)
	}
}
