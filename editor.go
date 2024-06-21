package main

import (
	"bufio"
	"fmt"
	"ollie/olliefile"
	"os"
	"strings"
)

func getWords(s *bufio.Scanner, o *olliefile.Ollie) error {
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
	}
	return nil
}

func parseCommand(c []string, o *olliefile.Ollie) (string, error) {
	cmdLen := len(c)
	if cmdLen > 2 {
		return nil, fmt.Errorf("Invalid command/parameters\n")
	}
	if c == "q" {
		if o.FileHandle != nil {
			o.FileHandle.Close()
		}
		os.Exit(0)
	}
	return fmt.Println("Unknown command"), nil
}

func main() {
	if len(os.Args) <= 1 || len(os.Args) > 2 {
		fmt.Println("ollie <file>")
		os.Exit(0)
	}

	of := olliefile.NewFile(os.Args[1])
	ws := bufio.NewScanner(os.Stdin)

	for {
		getWords(ws, of)
		fmt.Print("? ")
		ws.Scan()
		cmd := strings.Split(" ", ws.Text())
		resp, err := parseCommand(cmd)
	}
}
