package main

import (
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"io"
	"os"
)

// fileList is the structure wrapping task lists to be stored on-disk.
type fileList struct {
	Definite  []*DefiniteTask
	Eventual  []*EventualTask
	Recurring []*RecurringTaskGenerator
}

var (
	UnknownTaskType = errors.New("Task type unknown to file list encoder")
)

// ReadList decodes a JSON-encoded fileList from the given io.Reader,
// converts it to a List, then sorts and returns it.
func ReadList(r io.Reader) (fl fileList, err error) {
	err = json.NewDecoder(r).Decode(&fl)
	return fl, err
}

// Write JSON-encodes the fileList to the given io.Writer.
func (fl fileList) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(fl)
}

// ReadListFile wraps ReadList and returns a fileList. If the file
// given does not exist, then isNew will be true.
func ReadListFile(path string) (fl fileList, isNew bool, err error) {
	// Try to read the file. If the error is that the file doesn't
	// exist, return an empty list, or otherwise return an error.
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		glog.Infof("List file %q doesn't exist, using blank\n", path)
		return fl, true, nil
	} else if err != nil {
		return
	}
	defer f.Close()

	fl, err = ReadList(f)
	return fl, false, err
}

// WriteFile wraps Write to encode the fileList to a file.
func (fl fileList) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return fl.Write(f)
}

// List converts a fileList to a List, sorts it, and returns it.
func (fl fileList) List() (l List) {
	// Find the length, roughly, and make a List with that capacity.
	length := len(fl.Definite) + len(fl.Eventual)
	l = make(List, 0, length)

	// Loop through each field and append
	for _, t := range fl.Definite {
		l = append(l, t.Tasks()...)
	}
	for _, t := range fl.Eventual {
		l = append(l, t.Tasks()...)
	}
	for _, t := range fl.Recurring {
		l = append(l, t.Tasks()...)
	}

	l.Sort()
	return l
}
