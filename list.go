package main

import (
	"errors"
	"github.com/aybabtme/color"
	"github.com/golang/glog"
	"os"
	"sort"
	"time"
)

const (
	// DueFmt is the date format with which due dates are displayed in
	// stringified Tasks.
	DueFmt = "Monday, Jan 02, 15:04"

	ColorThreshold = time.Hour * 24
)

var (
	Rainbow = []color.Paint{
		color.RedPaint,
		color.YellowPaint,
		color.GreenPaint,
		color.CyanPaint,
		color.BluePaint,
		color.PurplePaint,
	}
)

var (
	InvalidTimeFormatError = errors.New("Invalid time format")
)

type List []Task

func ReadListFile(path string) (l List, err error) {
	// Try to read the file. If the error is that the file doesn't
	// exist, return an empty list, or otherwise return an error.
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		glog.Infof("List file %q doesn't exist, using blank\n", path)
		return List{}, nil
	} else if err != nil {
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

// ColorForDate returns a color.Brush appropriate for the given date,
// according to the given threshold. It goes in spectrum order from
// purple to red, as terminal colors allow, in order of increasing
// urgency.
func ColorForDate(dueby time.Time, threshold time.Duration) color.Brush {
	// Determine how far away the due date is.
	distance := dueby.Sub(time.Now())

	// Determine which paint to use by finding the number of times the
	// threshold goes into the distance.
	col := int(distance / threshold)

	// Store this for a moment so that we don't have to keep invoking
	// len().
	sizeRainbow := len(Rainbow)

	if col >= sizeRainbow {
		col = len(Rainbow) - 1
	} else if col < 0 {
		col = 0
	}

	// Return a new brush with the default background color and the
	// calculated foreground color.
	return color.NewBrush(color.Paint(""), Rainbow[col])
}
