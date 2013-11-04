package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aybabtme/color"
	"github.com/golang/glog"
	"io"
	"math"
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

// Nice calculates the numerical nice value for a Task, so that it can
// be sorted easily. The formula is
//
//     log(priority) * (due - now)
//
// where timestamps are UNIX dates.
func (t *Task) Nice() int {
	return int(
		math.Log(float64(t.Priority)) *
			float64(t.DueBy.Sub(time.Now())))
}

// String allows Tasks to be stringified easily. If the global Context
// specifies that color is allowed, it will be used.
func (t *Task) String() string {
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
func (t *Task) LongString() string {
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
