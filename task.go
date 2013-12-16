package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Task interface {
	// Nice is the integer value which determines how urgent the task
	// is, with lower values meaning greater urgency.
	Nice() int

	// Match checks whether a given search term should match the task,
	// usually comparing the task name, if appropriate. There is no
	// case guarantee.
	Match(string) bool

	// String formats the task in a brief list-friendly format,
	// typically without a description.
	String() string

	// LongString is similar to String, but includes the description
	// if appropriate.
	LongString() string
}

type DefiniteTask struct {
	Priority          int
	DueBy             time.Time
	Name, Description string
}

// Nice calculates the numerical nice value for a DefiniteTask, so
// that it can be sorted easily. The formula is
//
//     log(priority) * (due - now)
//
// where timestamps are UNIX dates.
func (t *DefiniteTask) Nice() int {
	return int(
		math.Log(float64(t.Priority)) *
			float64(t.DueBy.Sub(time.Now())))
}

// Match checks whether the given search term matches the task's title
// case-insensitively and returns the result.
func (t *DefiniteTask) Match(term string) bool {
	return strings.HasPrefix(
		strings.ToLower(t.Name), strings.ToLower(term))
}

// String allows DefiniteTasks to be stringified easily. If the global
// Context specifies that color is allowed, it will be used.
func (t *DefiniteTask) String() string {
	// Colorize if appropriate.
	if Ctx.Colors {
		// Find the appropriate color based on the imminence of the
		// due date.
		col := ColorForDate(t.DueBy, ColorThreshold)

		return fmt.Sprintf(
			col("(%d) %s - %s")+"\n",
			t.Priority, t.DueBy.Format(DueFmt), t.Name)
	}

	return fmt.Sprintf("(%d) %s - %s\n",
		t.Priority, t.DueBy.Format(DueFmt), t.Name)
}

// LongString allows Tasks to be stringified in full, including the
// description. Its behavior is similar to String.
func (t *DefiniteTask) LongString() string {
	// Colorize if appropriate.
	if Ctx.Colors {
		// Find the appropriate color based on the imminence of the
		// due date.
		col := ColorForDate(t.DueBy, ColorThreshold)

		return fmt.Sprintf(
			col("(%d) %s - %s")+"\n\t%s\n",
			t.Priority, t.DueBy.Format(DueFmt),
			t.Name, t.Description)
	}

	return fmt.Sprintf("(%d) %s - %s\n\t%s\n",
		t.Priority, t.DueBy.Format(DueFmt),
		t.Name, t.Description)
}
