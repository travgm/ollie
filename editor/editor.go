package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"git.sr.ht/~travgm/ollie/tree/develop/conf"
	"git.sr.ht/~travgm/ollie/tree/develop/olliefile"
	"git.sr.ht/~travgm/ollie/tree/develop/spellcheck"
)

func getWords(s *bufio.Scanner, o *olliefile.File) error {
	if s == nil {
		return fmt.Errorf("GetWords Error, Scanner is empty\n")
	}

	for s.Scan() {
		if s.Text() == "." {
			break
		}
		o.Lines = append(o.Lines, s.Text())
		o.LineCount += 1
		o.WordCount += len(strings.Split(" ", s.Text()))
		fmt.Println(o.LineCount)
	}
	return nil
}

func execCommand(c []string, s *bufio.Scanner, o *olliefile.File) (string, error) {
	cmdLen := len(c)
	if cmdLen > 2 {
		return "", fmt.Errorf("Invalid command/parameters\n")
	}
	cmd := c[0]
	param := ""
	if cmdLen > 1 {
		param = c[1]
	}
	
	switch cmd {
	case "q":
		if o.FileHandle != nil {
			o.FileHandle.Close()
		}
		os.Exit(0)		
	case "a":
		err := getWords(s, o)
		if err != nil {
			return "", err
		}
		return "", nil
	case "w":
		if param != "" {
			o.Name = param
			err := o.CreateFile()
			if err != nil {
				return "", err
			}
		}

		bytes, err := o.WriteFile()
		fmt.Printf("Wrote %d bytes to %s\n", bytes, o.Name)
		return string(bytes), err
	case "i":
		fmt.Println(o)
		return "", nil
	default:
		return "", fmt.Errorf("unknown command")			
	}
	return "", nil
}

func initEditor(args []string) (*conf.Settings, *olliefile.File) {
	of := &olliefile.File{Name: "junk.ollie"}
	if len(args) == 2 {
		of.Name = args[1]
		of.CreateFile()
	}
	config, err := conf.ParseConfig()
	if err != nil {
		return nil, of
	}

	_ = spellcheck.Dict{}

	return config, of
}

func main() {
	if len(os.Args) < 1 || len(os.Args) > 2 {
		fmt.Println("ollie <file>")
		os.Exit(0)
	}

	_, of := initEditor(os.Args)

	ws := bufio.NewScanner(os.Stdin)
	for {
		getWords(ws, of)
		fmt.Print("? ")
		ws.Scan()
		cmd := strings.Split(ws.Text(), " ")
		_, err := execCommand(cmd, ws, of)
		if err != nil {
			fmt.Println(err)
		}
	}
}
