package main

import (
	"github.com/aybabtme/color"
	"time"
)

const (
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

	// NoneBrush is a color.Brush which just returns the given string
	// without colorizing it.
	NoneBrush = color.Brush(func(s string) string { return s })
)

// BrushConditionally checks whether the given ctx requests
// colorization, and if so, returns the same brush, or otherwise
// returns a blank brush.
func BrushConditionally(ctx *Context, brush color.Brush) color.Brush {
	if ctx.Colors {
		return brush
	} else {
		return NoneBrush
	}
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
		col = sizeRainbow - 1
	} else if col < 0 {
		col = 0
	}

	// Return a new brush with the default background color and the
	// calculated foreground color.
	return color.NewBrush(color.Paint(""), Rainbow[col])
}

// ColorForPriority selects a color.Brush appropriate for the given
// priority, according to the given threshold. It follows the same
// guidelines as ColorForDate.
func ColorForPriority(priority int, threshold int) color.Brush {
	// Determine which paint to use by finding the number of times the
	// threshold goes into the distance.
	col := priority / threshold

	// Use this color if the index exists in the rainbow, or if not,
	// the last color in the rainbow.
	sizeRainbow := len(Rainbow)
	if col >= sizeRainbow {
		col = sizeRainbow - 1
	} else if col < 0 {
		col = 0
	}

	// Return a new brush with the default background color and the
	// calculated foreground color.
	return color.NewBrush(color.Paint(""), Rainbow[col])
}
