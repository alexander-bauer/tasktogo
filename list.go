package main

import (
	"errors"
	"sort"
)

const (
	// RelFmt and DueFmt are time format strings used for formatting
	// tasks in the list view. The former is used if the due date is
	// relatively nearby, and the latter is used if not.
	RelFmt = "%s at 15:04"
	DueFmt = "Jan 02, 15:04"
)

var (
	InvalidTimeFormatError = errors.New("Invalid time format")
)

type List []Task

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

	// Calculate the nice values. A lower value implies a higher
	// precedence.
	xnice, ynice := x.Nice(), y.Nice()

	// If the task at i (x) should be sorted before the one at j (y),
	// return true. Note that we use strictly less than and strictly
	// greater than, because if they are equal, we will attempt to
	// determine by another method.
	if xnice < ynice {
		return true
	} else {
		return false
	}
}

// Swap does a simple swap of two items. (For use with package sort.)
func (l List) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
