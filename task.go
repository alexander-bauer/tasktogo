package main

import (
	"fmt"
	"github.com/SashaCrofter/reltime"
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

	// Done is used to remove a Task from being displayed again after
	// it has been marked completed by the user.
	Done(*fileList)
}

type TaskContainer interface {
	// Tasks returns a representation of the TaskContainer as a slice
	// of Tasks, which may be nil.
	Tasks() []Task
}

const (
	// EventualFactor is the amount of time by which the priorities on
	// eventual tasks are multiplied.
	EventualFactor = float64(time.Hour * 72)

	// EventualThreshold is the about by which a priority value must
	// be increased in order for it to be colorized differently.
	EventualThreshold = 1
)

type DefiniteTask struct {
	Priority          int
	DueBy             time.Time
	Name, Description string
}

// Tasks causes DefiniteTask to satisfy the TaskContainer interface.
func (t *DefiniteTask) Tasks() []Task {
	return []Task{t}
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
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx, ColorForDate(t.DueBy, ColorThreshold))

	return fmt.Sprintf(col("(%d) %s - %s\n"),
		t.Priority, reltime.FormatRelative(RelFmt, DueFmt, t.DueBy), t.Name)
}

// LongString allows Tasks to be stringified in full, including the
// description. Its behavior is similar to String.
func (t *DefiniteTask) LongString() string {
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx, ColorForDate(t.DueBy, ColorThreshold))

	return fmt.Sprintf(col("(%d) %s - %s\n\t%s\n"),
		t.Priority, reltime.FormatRelative(RelFmt, DueFmt, t.DueBy),
		t.Name, t.Description)
}

func (t *DefiniteTask) Done(fl *fileList) {
	for i, container := range fl.Definite {
		if container == t {
			fl.Definite = append(fl.Definite[:i], fl.Definite[i+1:]...)
		}
	}
}

// EventualTask floats around in the todo list, remaining at a
// constant Nice value.
type EventualTask struct {
	Priority          int
	Name, Description string
}

// Tasks causes EventualTask to satisfy the TaskContainer interface.
func (t *EventualTask) Tasks() []Task {
	return []Task{t}
}

func (t *EventualTask) Nice() int {
	return int(math.Log(float64(t.Priority)) * EventualFactor)
}

func (t *EventualTask) Match(term string) bool {
	return strings.HasPrefix(
		strings.ToLower(t.Name), strings.ToLower(term))
}

func (t *EventualTask) String() string {
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx,
		ColorForPriority(t.Priority, EventualThreshold))

	return fmt.Sprintf(col("(%d) - %s\n"), t.Priority, t.Name)
}

func (t *EventualTask) LongString() string {
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx,
		ColorForPriority(t.Priority, EventualThreshold))

	return fmt.Sprintf(col("(%d) - %s\n\t%s\n"),
		t.Priority, t.Name, t.Description)
}

func (t *EventualTask) Done(fl *fileList) {
	for i, container := range fl.Eventual {
		if container == t {
			fl.Eventual = append(fl.Eventual[:i], fl.Eventual[i+1:]...)
		}
	}
}

// RecurringTaskGenerator is a generator tasks that occur at a regular
// interval.
type RecurringTaskGenerator struct {
	DoneUntil time.Time
	Except    []time.Time

	Start, End time.Time
	Delay      time.Duration

	// Spawn is a template for the generated RecurringTask with its
	// parent, occurrence counter, and due date unset. Its name and
	// description can optionally be printf format strings, which are
	// sprinted with the occurrence number (1-indexed) as the
	// argument.
	Spawn RecurringTask
}

// Tasks allows the RecurringTaskGenerator to produce all of its child
// tasks based on stored parameters.
func (g *RecurringTaskGenerator) Tasks() []Task {
	// Find the time to start generating from. If the Start time comes
	// before the DoneUntil marker, use the DoneUntil marker.
	start := g.Start
	if g.DoneUntil.After(start) {
		start = g.DoneUntil
	}

	// Find the number of tasks spawned since the start time by
	// finding the difference between the start time and the current
	// time, and dividing by the Delay.
	numBetween := int(time.Now().Sub(start) / g.Delay)
	if numBetween < 0 {
		numBetween = 0
	}

	// Also include the number of exceptions, and the next occurrence.
	numTasks := len(g.Except) + numBetween + 1
	tasks := make([]Task, 0, numTasks)

	for _, dueby := range g.Except {
		tasks = append(tasks,
			g.SpawnTask(
				int(dueby.Sub(g.Start)/g.Delay)+1, // occurrence
				dueby))
	}

	// Find the number of tasks that occurred before the start marker
	// marker, inclusive, so that we can get the occurrence number
	// correct.
	numBefore := int(start.Sub(g.Start)/g.Delay) + 1

	dueby := start.Add(time.Duration(numBefore) * g.Delay)
	for i := 0; i < numBetween+1; i++ {
		tasks = append(tasks,
			g.SpawnTask(
				i+numBefore,
				dueby))

		dueby = dueby.Add(g.Delay)
	}

	return tasks
}

func (g *RecurringTaskGenerator) SpawnTask(occurrence int,
	dueby time.Time) *RecurringTask {

	// Copy the Spawn and set the parent.
	newtask := g.Spawn
	newtask.parent = g

	// Set the fields from the arguments.
	newtask.Occurrence = occurrence
	newtask.DueBy = dueby

	// Sprintf the remaining fields.
	newtask.Name = fmt.Sprintf(g.Spawn.Name, occurrence)
	newtask.Description = fmt.Sprintf(g.Spawn.Description, occurrence)

	return &newtask
}

// Done modifies the state of the generator such that a Task with the
// given dueby date will not be produced again.
func (g *RecurringTaskGenerator) Done(dueby time.Time) {
	// TODO implement here
}

type RecurringTask struct {
	// parent is a pointer to the RecurringTaskGenerator that
	// generated this task.
	parent *RecurringTaskGenerator

	// Occurrence is the 1-indexed occurrence number of this task.
	Occurrence int

	Priority          int
	DueBy             time.Time
	Name, Description string
}

func (t *RecurringTask) Nice() int {
	return int(
		math.Log(float64(t.Priority)) *
			float64(t.DueBy.Sub(time.Now())))
}

// Match checks whether the given search term matches the task's title
// case-insensitively and returns the result.
func (t *RecurringTask) Match(term string) bool {
	return strings.HasPrefix(
		strings.ToLower(t.Name), strings.ToLower(term))
}

func (t *RecurringTask) String() string {
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx, ColorForDate(t.DueBy, ColorThreshold))

	return fmt.Sprintf(col("(%d) %s - %s\n"),
		t.Priority, reltime.FormatRelative(RelFmt, DueFmt, t.DueBy), t.Name)
}

func (t *RecurringTask) LongString() string {
	// Get a function for colorizing the string if appropriate. If
	// Ctx.Colors is not set, then it will do nothing.
	col := BrushConditionally(Ctx, ColorForDate(t.DueBy, ColorThreshold))

	return fmt.Sprintf(col("(%d) %s - %s\n\t%s\n"),
		t.Priority, reltime.FormatRelative(RelFmt, DueFmt, t.DueBy),
		t.Name, t.Description)
}

func (t *RecurringTask) Done(fl *fileList) {
	t.parent.Done(t.DueBy)
}
