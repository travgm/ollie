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

type Channels struct {
	executor chan string
	spelling chan []string
	spellres chan []string
	done     chan string
}

func execSpellchecker(channel Channels) {
	dict := spellcheck.Dict{WordFile: "/usr/share/dict/words", MaxSuggest: 3}
	err := dict.LoadWordlist()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		// We received a message to the spellchecker. We spell check the slice
		// and send back a slice that has suggestions.
		case words, ok := <-channel.spelling:
			if ok {
				suggestion := make([]string, 1)
				for _, word := range words {
					vals, err := dict.CheckWord(word)
					if err != nil || len(vals) < 1 {
						continue
					}
					suggestion = append(suggestion, vals...)
				}
				channel.spellres <- suggestion
			}
		case <-channel.done:
			return
		}
	}
}

func execCommand(channel Channels, s *bufio.Scanner, o *olliefile.File) {
	for {
		select {
		case val, ok := <-channel.executor:
			if ok {
				c := strings.Split(val, " ")

				cmdLen := len(c)
				if cmdLen > 2 {
					fmt.Errorf("Invalid command/parameters\n")
				}

				cmd := c[0]
				param := ""
				if cmdLen > 1 {
					param = c[1]
				}

				switch cmd {
				case "a":
					err := getWords(channel, s, o)
					if err != nil {
						fmt.Println(err)
					}
				case "w":
					if param != "" {
						o.Name = param
						err := o.CreateFile()
						if err != nil {
							fmt.Println(err)
						}
					}

					bytes, err := o.WriteFile()
					fmt.Printf("Wrote %d bytes to %s\n", bytes, o.Name)
					if err != nil {
						fmt.Println(err)
					}
				case "i":
					fmt.Println(o)
				default:
					fmt.Println("unknown command")
				}
			}
		case <-channel.done:
			return
		}
	}
}

func getWords(channel Channels, s *bufio.Scanner, o *olliefile.File) error {
	if s == nil {
		return fmt.Errorf("GetWords Error, Scanner is empty\n")
	}

	// We will eventually get this from the config
	shouldSpellcheck := true
	for s.Scan() {
		if s.Text() == "." {
			break
		}

		if shouldSpellcheck && len(s.Text()) >= 3 {
			channel.spelling <- strings.Fields(s.Text())
			fmt.Println("spellchecking...")
			val, ok := <-channel.spellres

			// Print suggestions
			count := 1
			if len(val) > 0 && ok {
				for _, suggest := range val {
					if suggest != "" {
						fmt.Printf(" %d:%s", count, suggest)
						count += 1
					}
				}
				if count == 1 {
					fmt.Printf("no suggestions\n")
				} else {
					fmt.Println("")
				}
			}
		}
		o.Lines = append(o.Lines, s.Text())
		o.LineCount += 1
		o.WordCount += len(strings.Split(" ", s.Text()))
		fmt.Println(o.LineCount)
	}
	return nil
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

	channels := Channels{
		executor: make(chan string, 1),
		spelling: make(chan []string, 1),
		spellres: make(chan []string, 1),
		done:     make(chan string, 1),
	}

	go execCommand(channels, ws, of)
	go execSpellchecker(channels)

	for {
		getWords(channels, ws, of)
		fmt.Print("? ")
		ws.Scan()
		in := ws.Text()
		if in == "q" {
			close(channels.done)
			break
		}
		channels.executor <- in
	}
}
