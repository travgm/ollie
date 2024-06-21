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
		return "", fmt.Errorf("Invalid command/parameters\n")
	}
	cmd := c[0]
	param := ""
	if cmdLen > 1 {
		param = c[1]
	}
	if cmd == "q" {
		if o.FileHandle != nil {
			o.FileHandle.Close()
		}
		os.Exit(0)
	} else if cmd == "w" {
		if param != "" {
			o.Name = param
			o.OCreateFile()
		}
		if o.FileHandle != nil && o.Name != "" {
			bytes, err := o.OWriteFile()
			fmt.Printf("Wrote %d bytes to %s\n", bytes, o.Name)
			return string(bytes), err
		} else {
			return "", fmt.Errorf("ERROR: Supply file name after command: w")
		}
	}

	return "", fmt.Errorf("Unknown Command")
}

func main() {
	if len(os.Args) < 1 || len(os.Args) > 2 {
		fmt.Println("ollie <file>")
		os.Exit(0)
	}

	of := olliefile.ONewFile("junk.ollie")
	if len(os.Args) == 2 {
		of.Name = os.Args[1]
		of.OCreateFile()
	}
	ws := bufio.NewScanner(os.Stdin)

	for {
		getWords(ws, of)
		fmt.Print("? ")
		ws.Scan()
		cmd := strings.Split(ws.Text(), " ")
		_, err := parseCommand(cmd, of)
		if err != nil {
			fmt.Println(err)
		}
	}
}
