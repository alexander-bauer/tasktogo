package main

import (
	"fmt"
	"github.com/gobs/args"
)

const (
	PromptString = ": "
)

// Prompt writes a prompt to the screen and reads the Input stream
// until a linebreak or EOF, then parses that into a Command type. If
// it encounters an error, no the command that's been parsed so far
// will be returned. It will not print any more than the initial
// prompt.
func Prompt(ctx *Context) (c *Command, err error) {
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
	// backslashes, and parse them into a command immediately.
	c, err = ParseCommand(args.GetArgs(line))

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
