// MIT License
//
// # Copyright (c) 2024 Travis Montoya and the ollie contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// olliefile holds the Ollie file struct with the information needed for
// handling file routines.
package olliefile

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type File struct {
	Name       string
	FileHandle *os.File
	Lines      []string
	WordCount  int
	LineCount  int
	Saved      bool
	LastSaved  time.Time
}

func (o *File) String() string {
	return fmt.Sprintf("File: %s\nLine Count: %d\nWord Count: %d\nLast Saved: %s",
		o.Name, o.LineCount, o.WordCount, o.LastSaved.Format("2006-01-02 15:04:05"))
}

func (o *File) WriteFile() (int, error) {
	if o.FileHandle == nil {
		return 0, fmt.Errorf("Error: File handle null")
	}

	err := o.FileHandle.Truncate(0)
	if err != nil {
		return 0, err
	}

	_, err = o.FileHandle.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	bytes := 0
	for _, s := range o.Lines {
		bw, err := fmt.Fprintln(o.FileHandle, s)
		bytes += bw
		if err != nil {
			return bytes, err
		}
	}
	o.FileHandle.Sync()
	o.Saved = true
	o.LastSaved = time.Now()
	return bytes, nil
}

func (o *File) UpdateLine(lineNumber string, str string) error {
	line, err := strconv.ParseInt(lineNumber, 10, 32)
	if err != nil {
		return fmt.Errorf("param", err)
	}

	if line < 1 || line > int64(len(o.Lines)) {
		return fmt.Errorf("invalid line number")
	}

	o.Lines[line-1] = str

	err = o.FileHandle.Truncate(0)
	if err != nil {
		return err
	}

	_, err = o.FileHandle.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = o.WriteFile()
	return err
}

func (o *File) CreateFile() error {
	if o.Name == "" {
		return fmt.Errorf("ERROR: No file name speified")
	}

	doesExist, err := os.OpenFile(o.Name, os.O_RDONLY, 0)
	if err == nil {
		o.FileHandle = doesExist
		err := o.readFile()
		if err != nil {
			return fmt.Errorf("read file", err)
		}
	}

	f, err := os.OpenFile(o.Name, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	o.FileHandle = f
	return nil
}

func (o *File) readFile() error {
	defer o.FileHandle.Close()

	scanner := bufio.NewScanner(o.FileHandle)
	for scanner.Scan() {
		o.Lines = append(o.Lines, scanner.Text())
		o.LineCount += 1
		o.WordCount += len(strings.Split(" ", scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	o.Saved = true
	o.LastSaved = time.Now()

	return nil
}
