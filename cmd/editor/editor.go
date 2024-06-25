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
			if err != nil {
				fmt.Println(err)
			}
		case <-done:
			return
		}
	}
}

func execSpellchecker(spelling chan []string, done <-chan string) {
	dict := spellcheck.Dict{WordFile: "/usr/share/dict/words", MaxSuggest: 3}
	err := dict.LoadWordlist()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		// We received a message to the spellchecker. We spell check the slice
		// and send back a slice that has suggestions.
		case words := <-spelling:
			suggestion := make([]string, dict.MaxSuggest)
			for _, word := range words {
				vals, err := dict.CheckWord(word)
				if err != nil {
					fmt.Println(err)
					continue
				}
				suggestion = append(suggestion, vals...)
			}
			spelling <- suggestion
		case <-done:
			return
		}
	}
}

func getWords(spelling chan []string, s *bufio.Scanner, o *olliefile.File) error {
	if s == nil {
		return fmt.Errorf("GetWords Error, Scanner is empty\n")
	}

	// We will eventually get this from the config
	shouldSpellcheck := true
	for s.Scan() {
		if s.Text() == "." {
			break
		}

		// set shouldSpellcheck to true to turn on spellchecking after each entry
		// currently there is a bug printing ,,,,,,,word1,,,,,word2 etc...
		if shouldSpellcheck && len(s.Text()) >= 3 {
			spelling <- strings.Fields(s.Text())
			fmt.Println("spellchecking...")
			val := <-spelling

			fmt.Printf("suggestions: %s\n", strings.Join(val, ", "))
		}
		o.Lines = append(o.Lines, s.Text())
		o.LineCount += 1
		o.WordCount += len(strings.Split(" ", s.Text()))
		fmt.Println(o.LineCount)
	}
	return nil
}

func execCommand(executor <-chan string, receiver chan<- error,
	done <-chan string, spelling chan []string, s *bufio.Scanner, o *olliefile.File) {
	for {
		select {
		case val := <-executor:
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
				err := getWords(spelling, s, o)
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
		case <-done:
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

	executor := make(chan string, 1)   // executes the instructions
	receiver := make(chan error, 1)    // receives errors, error handling
	spelling := make(chan []string, 1) // spellchecker

	// channel to signify completion
	done := make(chan string, 1)

	// TODO: Move channels to a struct? so we can pass the struct around instead of
	// way to many channels
	go execCommand(executor, receiver, done, spelling, ws, of)
	go execSpellchecker(spelling, done)
	go stateReceiver(receiver, done)

	for {
		getWords(spelling, ws, of)
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
