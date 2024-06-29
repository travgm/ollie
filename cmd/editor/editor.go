package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"git.sr.ht/~travgm/ollie/conf"
	"git.sr.ht/~travgm/ollie/olliefile"
	"git.sr.ht/~travgm/ollie/spellcheck"
	"git.sr.ht/~travgm/ollie/version"
)

// Editor state, this should hold everything we need to be passed around
// and modify state. The only channels we use are a send and receive one
// for spellchecking and a done channel to signify the program is exiting
//
// command should only be one of the valid editor const commands
type State struct {
	channels  spellcheck.Channels
	command   string
	wordInput *bufio.Scanner
	ollie     *olliefile.File
	conf      *conf.Settings
}

// Editor commands
const (
	WRITE_FILE    = "w"
	APPEND        = "a"
	FILE_INFO     = "i"
	SPELLCHECK    = "p"
	FIX_LINE      = "f"
	EXEC_CMD      = "e"
	QUIT_EDITOR   = "q"
	DEL_LAST_LINE = "d"
	COMMAND_MODE  = "."
)

// Checks state.command and runs the proper routines for it
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
		break
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
	case DEL_LAST_LINE:
		line := strconv.Itoa(len(state.ollie.Lines))
		if state.ollie.FileHandle != nil {
			err := state.ollie.UpdateLine(line, "")
			if err != nil {
				fmt.Println(err)
			}
		}
		state.ollie.Lines = state.ollie.Lines[:len(state.ollie.Lines)-1]
		state.ollie.LineCount -= 1
		fmt.Println("cleared line", line)
	case FIX_LINE:
		state.wordInput.Scan()
		err := state.ollie.UpdateLine(param, state.wordInput.Text())
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("updated line", param)
		}
	case EXEC_CMD:
		if len(c) <= 1 {
			fmt.Println("'e' needs a parameter to run in the shell")
			break
		}
		scmd := exec.Command(c[1], c[2:]...)
		res, err := scmd.CombinedOutput()
		if err != nil {
			fmt.Println("exec failed:", err)
		} else {
			fmt.Println(string(res))
		}
	case WRITE_FILE:
		if param != "" {
			state.ollie.Name = param
			err := state.ollie.CreateFile()
			if err != nil {
				fmt.Println(err)
			}
		} else if state.ollie.FileHandle == nil && state.ollie.Name != "" {
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
	default:
		fmt.Println("unknown command")
	}
}

// Spellchecking a single word
//
// This sends the current text in the state bufio scanner to the spelling
// channel. It receives back single or multiple suggestions for each word
// in the string sent
func getSpellcheckSuggestions(state *State) error {
	state.channels.Spelling <- strings.Fields(state.wordInput.Text())
	fmt.Println("spellchecking...")
	val, ok := <-state.channels.Spellres
	if !ok {
		return fmt.Errorf("spellcheck channel closed")
	}
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

	return nil
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
			err := getSpellcheckSuggestions(state)
			if err != nil {
				fmt.Println(err)
			}
		}

		state.ollie.Lines = append(state.ollie.Lines, state.wordInput.Text())
		state.ollie.LineCount += 1
		state.ollie.WordCount += len(strings.Split(" ", state.wordInput.Text()))
		fmt.Printf("%d:%d\n", state.ollie.LineCount, len(state.wordInput.Text()))
	}
	return nil
}

func initEditor(filename string, spell bool) (State, error) {
	of := &olliefile.File{Name: "junk.ollie"}
	if filename != "" {
		of.Name = filename
		of.CreateFile()
	}
	config, err := conf.FromFile(conf.DefaultConfFile)
	if !errors.Is(err, os.ErrNotExist) {
		return State{}, nil
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
		return err
	}

	go spellcheck.ExecSpellchecker(state.channels)

	for {
		err := getWords(&state)
		if err != nil {
			fmt.Println(err)
			close(state.channels.Done)
			return err
		}
		fmt.Print("@ ")
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
