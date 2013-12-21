package main

import (
	"encoding/json"
	"errors"
	"io"
)

// fileList is the structure wrapping task lists to be stored
// on-disk. It provides its own JSON marshallers and unmarshallers.
type fileList struct {
	Definite []*DefiniteTask
	Eventual []*EventualTask
}

var (
	UnknownTaskType = errors.New("Task type unknown to file list encoder")
)

// ReadList decodes a JSON-encoded fileList from the given io.Reader,
// converts it to a List, then sorts and returns it.
func ReadList(r io.Reader) (l List, err error) {
	fl := fileList{}
	err = json.NewDecoder(r).Decode(&fl)
	if err != nil {
		return nil, err
	}
	return fl.List(), nil
}

// Write JSON-encodes the List to the given io.Writer by first
// converting it to a fileList. It does not sort the List.
func (l List) Write(w io.Writer) error {
	fl, err := toFileList(l)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(fl)
}

// List converts a fileList to a List, sorts it, and returns it.
func (fl *fileList) List() (l List) {
	l = make(List, len(fl.Definite)+len(fl.Eventual))

	// Now, loop through each individual element of the fileList,
	// convert it to the Task interface, and place it in the list.
	j := 0
	for _, t := range fl.Definite {
		l[j] = Task(t)
		j++
	}
	for _, t := range fl.Eventual {
		l[j] = Task(t)
		j++
	}

	l.Sort()
	return l
}

func toFileList(l List) (fl *fileList, err error) {
	fl = &fileList{
		Definite: make([]*DefiniteTask, 0),
		Eventual: make([]*EventualTask, 0),
	}

	for _, t := range l {
		if converted, ok := t.(*DefiniteTask); ok {
			fl.Definite = append(fl.Definite, converted)
		} else if converted, ok := t.(*EventualTask); ok {
			fl.Eventual = append(fl.Eventual, converted)
		} else {
			return nil, UnknownTaskType
		}
	}

	return
}
