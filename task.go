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
	// LastCompleted marks the most recent task ID (1-indexed) to have
	// been marked complete.
	LastCompleted int
	// Except is a slice containing all task IDs that are less than
	// LastCompleted, but which have *not* been marked complete.
	Except []int

	Start, End time.Time
	Delay      []time.Duration

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
	// Find the last task ID that will be generated.

	// If the current time is less than the End time, add one extra
	// task for the one currently in session, as long as the session
	// has actually started.
	var finalID int
	endTime := time.Now()
	if !g.End.IsZero() && endTime.After(g.End) {
		endTime = g.End
	} else if !g.Start.After(endTime) {
		finalID++
	}

	// Find the number of instances of the delays that have happened
	// by the time between g.Start and current time.
	finalID += g.FindLastID(endTime)

	tasks := make([]Task, 0, finalID-g.LastCompleted+len(g.Except))

	// Add all the exceptions.
	for _, id := range g.Except {
		tasks = append(tasks, g.SpawnTask(id))
	}

	// Add every task since the latest one that's been marked
	// completed.
	for id := g.LastCompleted; id < finalID; id++ {
		tasks = append(tasks, g.SpawnTask(id+1))
	}

	return tasks
}

func (g *RecurringTaskGenerator) SpawnTask(occurrence int) *RecurringTask {
	// Copy the Spawn and set the parent.
	newtask := g.Spawn
	newtask.parent = g

	// Set the fields from the arguments.
	newtask.Occurrence = occurrence
	newtask.DueBy = g.DueByID(occurrence)

	// Sprintf the remaining fields.
	newtask.Name = fmt.Sprintf(g.Spawn.Name, occurrence)
	newtask.Description = fmt.Sprintf(g.Spawn.Description, occurrence)

	return &newtask
}

func (g *RecurringTaskGenerator) DueByID(occurrence int) time.Time {
	// Calculate the number of full turnovers of the delay schedule
	// the occurence is at, as well as its progress into the current
	// one.
	occurrence -= 1
	if occurrence < 0 {
		return time.Time{}
	}
	full, remaining := occurrence/len(g.Delay), occurrence%len(g.Delay)

	return g.Start.Add(time.Duration(full)*g.SumDelay(len(g.Delay)) +
		g.SumDelay(remaining))
}

func (g *RecurringTaskGenerator) SumDelay(lastIndex int) time.Duration {
	var sum time.Duration
	for _, delay := range g.Delay[:lastIndex] {
		sum += delay
	}
	return sum
}

func (g *RecurringTaskGenerator) FindLastID(t time.Time) (id int) {
	// First, find the number of complete delay schedule rollovers
	// there have been, and multiply that by the number of tasks in
	// the schedule. We add one because it's 1-indexed.
	id = int(t.Sub(g.Start)/g.SumDelay(len(g.Delay))) + 1

	// Next, find the number of tasks that have occurred in the
	// current schedule by looping through each and finding the first
	// one that occurs after the given time.
	taskTime := g.DueByID(id)
	for i, delay := range g.Delay {
		taskTime = taskTime.Add(delay)
		if taskTime.After(t) {
			// Add the index of the item in the schedule.
			id += i
			break
		}
	}

	return id
}

// Done modifies the state of the generator such that a Task with the
// ID date will not be produced again.
func (g *RecurringTaskGenerator) Done(id int) {
	// In the simplest case, the id is greater than the last completed
	// task, so the counter can simply be incremented.
	if id > g.LastCompleted {
		// For each ID inbetween the new ID and the last completed
		// one, add it to the list of exceptions.
		for i := g.LastCompleted + 1; i < id; i++ {
			g.Except = append(g.Except, i)
		}
		// Set the LastCompleted marker.
		g.LastCompleted = id

	} else {
		// If it is not greater than the last completed ID, then it
		// should be considered to be an exception, so locate and
		// remove it from the list.
		for i, exceptionID := range g.Except {
			if exceptionID == id {
				g.Except = append(g.Except[:i], g.Except[i+1:]...)
			}
		}
	}
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
	t.parent.Done(t.Occurrence)
}
