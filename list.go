package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

const (
	// DueFmt is the date format with which due dates are displayed in
	// stringified Tasks.
	DueFmt = "Monday, Jan 02, 15:04"
)

var (
	InvalidTimeFormatError = errors.New("Invalid time format")
)

type List []*Task

type Task struct {
	Priority          int
	DueBy             time.Time
	Name, Description string
}

// ReadList decodes a JSON-encoded List from the given io.Reader, then
// sorts and returns it.
func ReadList(r io.Reader) (l List, err error) {
	l = List{}
	err = json.NewDecoder(r).Decode(&l)
	if err != nil {
		return
	}
	l.Sort()
	return
}

// Write JSON-encodes the List to the given io.Writer, using
// non-pretty formatting. It does not sort the List.
func (l List) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(l)
}

func ReadListFile(path string) (l List, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	return ReadList(f)
}

func (l List) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return l.Write(f)
}

// Sort is a convenience function that invokes sort.Sort() on the
// given List.
func (l List) Sort() {
	sort.Sort(l)
}

// Len returns the len() of the List. (For use with package sort.)
func (l List) Len() int {
	return len(l)
}

// Less returns whether the Task at index i has a lower nice value
// (priority * (due date - now)) than that at index j. (For use with
// package sort.)
func (l List) Less(i, j int) bool {
	// Retrieve both values so that they don't have to be looked up
	// again.
	x, y := l[i], l[j]

	// Calculate the nice values, which are just priority * (due date
	// - now). A lower value implies a higher precedence.
	now := time.Now()
	xnice, ynice := x.Priority*int((x.DueBy.Sub(now))),
		y.Priority*int((y.DueBy.Sub(now)))

	// If the task at i (x) should be sorted before the one at j (y),
	// return true. Note that we use strictly less than and strictly
	// greater than, because if they are equal, we will attempt to
	// determine by another method.
	if xnice < ynice {
		return true
	} else if xnice > ynice {
		return false
	} else {
		// If they're equal, sort by due date. Note that if the due
		// dates are equal, the priorities must also be equal, and
		// they'll just end up being sorted by insertion order.
		if y.DueBy.After(x.DueBy) {
			return true
		} else {
			return false
		}
	}
}

// Swap does a simple swap of two items. (For use with package sort.)
func (l List) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// String allows Tasks to be stringified easily.
func (t *Task) String() string {
	return fmt.Sprintf("(%d) %s - %s\n\t%s\n",
		t.Priority, t.DueBy.Format(DueFmt),
		t.Name, t.Description)
}

type Time time.Time

func (t *Time) UnmarshalJSON(b []byte) error {
	// Remove quotes if possible, and otherwise, error.
	if len(b) > 2 {
		b = b[1 : len(b)-1]
	} else {
		return InvalidTimeFormatError
	}
	print(b)

	// Next, try to parse the time.
	newtime, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		return err
	}

	// Do some magic.
	*t = *(*Time)(&newtime)
	return nil
}
