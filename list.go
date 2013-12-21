package main

import (
	"errors"
	"github.com/golang/glog"
	"os"
	"sort"
)

const (
	// DueFmt is the date format with which due dates are displayed in
	// stringified Tasks.
	DueFmt = "Monday, Jan 02, 15:04"
)

var (
	InvalidTimeFormatError = errors.New("Invalid time format")
)

type List []Task

// ReadListFile wraps ReadList and returns a List decoded from a
// JSON-encoded fileList type. If the file given does not exist, then
// isNew will be true.
func ReadListFile(path string) (l List, isNew bool, err error) {
	// Try to read the file. If the error is that the file doesn't
	// exist, return an empty list, or otherwise return an error.
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		glog.Infof("List file %q doesn't exist, using blank\n", path)
		return List{}, true, nil
	} else if err != nil {
		return
	}
	defer f.Close()

	l, err = ReadList(f)
	return l, false, err
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
