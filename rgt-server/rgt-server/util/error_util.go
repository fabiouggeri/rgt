package util

import (
	"runtime"
	"strconv"
	"strings"
)

type StackEntry struct {
	file string
	line int
}

type StackError struct {
	entries []*StackEntry
}

func (e *StackEntry) File() string {
	return e.file
}

func (e *StackEntry) Line() int {
	return e.line
}

func FullStack() *StackError {
	return Stack(0)
}

func Stack(limit int) *StackError {
	stack := &StackError{entries: make([]*StackEntry, 0)}
	newLimit := limit + 2
	skip := 2
	for skip <= newLimit || limit <= 0 {
		if _, file, line, ok := runtime.Caller(skip); ok {
			stack.entries = append(stack.entries, &StackEntry{file: file, line: line})
		} else {
			return stack
		}
		skip++
	}
	return stack
}

func (s *StackError) Items() []*StackEntry {
	return s.entries
}

func (s *StackError) Length() int {
	return len(s.entries)
}

func (s *StackError) Get(index int) *StackEntry {
	if index >= 0 && index < len(s.entries) {
		return s.entries[index]
	}
	return nil
}

func (s *StackError) String() string {
	return s.ToString(0)
}

func (s *StackError) ToString(margin int) string {
	var sb strings.Builder
	for _, e := range s.entries {
		sb.WriteString(strings.Repeat(" ", margin))
		sb.WriteString(e.file)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(e.line))
		sb.WriteRune('\n')
	}
	return sb.String()
}
