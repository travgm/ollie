package olliefile

import (
	"fmt"
	"os"
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
	return fmt.Sprintf("File: %s\nLine Count: %d\nWord Count: %d\nLast Saved: %s\n",
		o.Name, o.LineCount, o.WordCount, o.LastSaved.Format("2006-01-02 15:04:05"))
}

func (o *File) WriteFile() (int, error) {
	if o.FileHandle == nil {
		return 0, fmt.Errorf("Error: File handle null")
	}
	bytes := 0
	for _, s := range o.Lines {
		bw, err := fmt.Fprintln(o.FileHandle, s)
		bytes += bw
		if err != nil {
			return bytes, err
		}
	}
	o.Saved = true
	o.LastSaved = time.Now()
	return bytes, nil
}

func (o *File) UpdateLine(line int64, str string) error {
	if line < 1 || line > int64(len(o.Lines)) {
		return fmt.Errorf("invalid line number")
	}

	o.Lines[line-1] = str

	err := o.FileHandle.Truncate(0)
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

	f, err := os.OpenFile(o.Name, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	o.FileHandle = f
	return nil
}
