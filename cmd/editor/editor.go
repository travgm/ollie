package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.sr.ht/~travgm/ollie/tree/develop/conf"
	"git.sr.ht/~travgm/ollie/tree/develop/olliefile"
	"git.sr.ht/~travgm/ollie/tree/develop/spellcheck"
	"git.sr.ht/~travgm/ollie/tree/develop/version"
)

type State struct {
	channels  spellcheck.Channels
	command   string
	wordInput *bufio.Scanner
	ollie     *olliefile.File
	conf      *conf.Settings
}

// Editor commands
const (
	WRITE_FILE   = "w"
	APPEND       = "a"
	FILE_INFO    = "i"
	SPELLCHECK   = "p"
	FIX_LINE     = "f"
	QUIT_EDITOR  = "q"
	COMMAND_MODE = "."
)

func runCommand(state *State) {
	c := strings.Split(state.command, " ")

	cmdLen := len(c)
	if cmdLen > 2 {
		fmt.Errorf("Invalid command/parameters\n")
	}

	// Holds the main command to be executed
	cmd := c[0]

	param := ""
	if cmdLen > 1 {
		param = c[1]
	}

	switch cmd {
	case APPEND:
		// We just break here because how the main loop is designed it drops us back to the getWords() function
		break
	case WRITE_FILE:
		if param != "" {
			state.ollie.Name = param
			err := state.ollie.CreateFile()
			if err != nil {
				fmt.Println(err)
			}

			// If the user doesn't specify a file to write to we write it to the junk file
		} else {
			err := state.ollie.CreateFile()
			if err != nil {
				fmt.Println(err)
			}
		}

		bytes, err := state.ollie.WriteFile()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Wrote %d bytes to %s\n", bytes, state.ollie.Name)
		}
	case FILE_INFO:
		fmt.Println(state.ollie)
	case SPELLCHECK:
		switch param {
		case "on":
			state.channels.ShouldSpellcheck = true
		case "off":
			state.channels.ShouldSpellcheck = false
		default:
			fmt.Println("valid parameter for spellcheck is 'on' or 'off'")
		}
	case FIX_LINE:
		i, err := strconv.ParseInt(param, 10, 32)
		if err != nil {
			fmt.Printf("param must be a valid line number\n")
		} else {
			state.wordInput.Scan()
			err := state.ollie.UpdateLine(i, state.wordInput.Text())
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("updated line %d\n", i)
			}
		}
	default:
		fmt.Println("unknown command")
	}
}

func getWords(state *State) error {
	if state == nil {
		return fmt.Errorf("GetWords Error. State is null\n")
	}

	for state.wordInput.Scan() {
		if state.wordInput.Text() == COMMAND_MODE {
			break
		}

		if state.channels.ShouldSpellcheck && len(state.wordInput.Text()) >= 3 {
			state.channels.Spelling <- strings.Fields(state.wordInput.Text())
			fmt.Println("spellchecking...")
			val, ok := <-state.channels.Spellres

			count := 1
			if len(val) > 0 && ok {
				fmt.Printf("corrections:")
				for _, suggest := range val {
					if suggest != "" {
						fmt.Printf(" %d:%s", count, suggest)
						count += 1
					}
				}
				if count == 1 {
					fmt.Printf(" no suggestions\n")
				} else {
					fmt.Println("")
				}
			}
		}
		state.ollie.Lines = append(state.ollie.Lines, state.wordInput.Text())
		state.ollie.LineCount += 1
		state.ollie.WordCount += len(strings.Split(" ", state.wordInput.Text()))
		fmt.Println(state.ollie.LineCount)
	}
	return nil
}

func initEditor(filename string, spell bool) (State, error) {
	of := &olliefile.File{Name: "junk.ollie"}
	if filename != "" {
		of.Name = filename
		of.CreateFile()
	}
	config, err := conf.ParseConfig()
	if err != nil {
		return State{}, err
	}

	spChannels := spellcheck.Channels{
		ShouldSpellcheck: spell,
		Spelling:         make(chan []string, 1),
		Spellres:         make(chan []string, 1),
		Done:             make(chan string, 1),
	}

	state := State{
		channels:  spChannels,
		wordInput: bufio.NewScanner(os.Stdin),
		ollie:     of,
		conf:      config,
	}

	return state, nil
}

func printUsage() {
	fmt.Println("Usage: ollie <file>")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	aboutFlag := flag.Bool("version", false, "Display version information")
	spellFlag := flag.Bool("spcheck", false, "Turn spellchecking on, default is off")

	flag.Parse()
	if *aboutFlag {
		version.DisplayAbout()
		return nil
	}

	if flag.NArg() > 1 {
		printUsage()
		return fmt.Errorf("incorrect number of arguments")

	}

	state, err := initEditor(flag.Arg(0), *spellFlag)
	if err != nil {
		fmt.Println(err)
		return err
	}

	go spellcheck.ExecSpellchecker(state.channels)

	for {
		err := getWords(&state)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Print("? ")
		state.wordInput.Scan()
		state.command = state.wordInput.Text()
		if state.command == QUIT_EDITOR {
			close(state.channels.Done)
			break
		}
		runCommand(&state)
	}

	return nil
}
