package olliefile

import (
	"fmt"
	"os"
	"time"
)

type Ollie struct {
	Name       string
	FileHandle *os.File
	Lines      []string
	WordCount  int
	LineCount  int
	Saved      bool
	LastSaved  time.Time
}

func ONewFile(name string) *Ollie {
	return &Ollie{Name: name, WordCount: 0, LineCount: 0, Saved: false, FileHandle: nil}
}

func (o *Ollie) OWriteFile() (int, error) {
	if o.FileHandle == nil {
		return 0, fmt.Errorf("Error: File handle null")
	}
	bytes := 0
	for _, s := range o.Lines {
		bw, err := fmt.Fprintln(o.FileHandle, s)
		if err != nil {
			return 0, err
		}
		bytes += bw
	}
	o.Saved = true
	o.LastSaved = time.Now()
	return bytes, nil
}

func (o *Ollie) OCreateFile() error {
	if o.Name == "" {
		return fmt.Errorf("ERROR: No file name speified")
	}

	f, err := os.OpenFile(o.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	o.FileHandle = f
	return nil
}
