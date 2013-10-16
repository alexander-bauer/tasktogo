package main

import (
	"encoding/json"
	"errors"
	"io"
	"time"
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

func ReadList(r io.Reader) (l List, err error) {
	l = List{}
	err = json.NewDecoder(r).Decode(&l)
	return
}

func (l List) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(l)
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
