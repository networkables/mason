// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package stackerr

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const maxdepth = 20

type StackErr struct {
	err       error
	stacktext string
	Frames    []Frame
}

func New(e error) StackErr {
	stack := make([]uintptr, maxdepth)
	depth := runtime.Callers(2, stack[:])
	trimstack := stack[:depth]
	frames := make([]Frame, len(trimstack))

	for i, pc := range trimstack {
		frames[i] = newFrame(pc)
	}
	buf := bytes.Buffer{}
	for _, frame := range frames {
		buf.WriteString(frame.String())
	}

	return StackErr{
		err:       e,
		stacktext: string(buf.Bytes()),
		Frames:    frames,
	}
}

func (s StackErr) Error() string {
	return s.err.Error()
}

func (s StackErr) Unwrap() error {
	return s.err
}

func (s StackErr) Stack() string {
	return s.stacktext
}

type Frame struct {
	File     string
	Line     int
	Function string
	Package  string
	PCStr    string
	Source   string
}

func newFrame(pc uintptr) Frame {
	f := Frame{PCStr: fmt.Sprintf("0x%x", pc)}
	rf := runtime.FuncForPC(pc)
	if rf != nil {
		f.Package, f.Function = packageAndName(rf)
		f.File, f.Line = rf.FileLine(pc - 1)
		f.Source, _ = f.sourceLine()
	}

	return f
}

func (f Frame) String() string {
	str := fmt.Sprintf("%s:%d (%s)\n", f.File, f.Line, f.PCStr)

	source, err := f.sourceLine()
	if err != nil {
		return str
	}

	return str + fmt.Sprintf("\t%s: %s\n", f.Function, source)
}

func (f Frame) sourceLine() (string, error) {
	if f.Line <= 0 {
		return "???", nil
	}

	file, err := os.Open(f.File)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == f.Line {
			return string(bytes.Trim(scanner.Bytes(), " \t")), nil
		}
		currentLine++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "???", nil
}

func packageAndName(fn *runtime.Func) (string, string) {
	name := fn.Name()
	pkg := ""

	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		pkg += name[:lastslash] + "/"
		name = name[lastslash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}

	name = strings.Replace(name, "·", ".", -1)
	return pkg, name
}
