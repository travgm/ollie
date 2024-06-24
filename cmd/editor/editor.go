package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"git.sr.ht/~travgm/ollie/tree/develop/conf"
	"git.sr.ht/~travgm/ollie/tree/develop/olliefile"
	"git.sr.ht/~travgm/ollie/tree/develop/spellcheck"
)

func stateReceiver(receiver <-chan error, done <-chan string) {
	for {
		select {
		case err := <-receiver:
			fmt.Println(err)
		case <-done:
			return
		}
	}
}

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

func execCommand(executor <-chan string, receiver chan<- error,
	done <-chan string, s *bufio.Scanner, o *olliefile.File) {
	for {
		select {
		case val := <- executor:
		c := strings.Split(val, " ")

		cmdLen := len(c)
		if cmdLen > 2 {
			receiver <- errors.New("Invalid command/parameters\n")
		}

		cmd := c[0]
		param := ""
		if cmdLen > 1 {
			param = c[1]
		}

		switch cmd {
		case "a":
			err := getWords(s, o)
			if err != nil {
				receiver <- err
			}
			receiver <- nil
		case "w":
			if param != "" {
				o.Name = param
				err := o.CreateFile()
				if err != nil {
					receiver <- err
				}
			}

			bytes, err := o.WriteFile()
			fmt.Printf("Wrote %d bytes to %s\n", bytes, o.Name)
			receiver <- err
		case "i":
			fmt.Println(o)
			receiver <- nil
		default:
			receiver <- errors.New("unknown command")
		}
		case <- done:
			return
		}
	}
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

	cf, of := initEditor(os.Args)
	if cf != nil {
		fmt.Println(cf)
	}

	ws := bufio.NewScanner(os.Stdin)

	executor := make(chan string, 1)
	receiver := make(chan error, 1)
	done := make(chan string, 1)
	go execCommand(executor, receiver, done, ws, of)
	go stateReceiver(receiver, done)

	for {
		getWords(ws, of)
		fmt.Print("? ")
		ws.Scan()
		in := ws.Text()
		if in == "q" {
			close(done)
			break
		}
		executor <- in
	}
}
