package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"

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
	SEARCH_TEXT   = "s"
	COMMAND_MODE  = "."
)

// Checks state.command and runs the proper routines for it
func execIoCommand(state *State) {

	cmd, param, err := parseCommandArgs(state)
	if err != nil {
		fmt.Errorf("Invalid command/parameters\n")
		return
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
		line, err := deleteLastLine(state)
		if err != nil {
			fmt.Println("error deleting last line", err)
		} else {
			fmt.Println("cleared line", line)
		}
	case FIX_LINE:
		state.wordInput.Scan()
		err := state.ollie.UpdateLine(param, state.wordInput.Text())
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("updated line", param)
		}
	case SEARCH_TEXT:
		_, err := searchLinesBuffer(state, param)
		if err != nil {
			fmt.Println("'s' needs a text string to search the buffer for")
		}
	case EXEC_CMD:
		res, err := shellCommand(param)
		if err != nil {
			fmt.Println("'e' error", err)
		} else {
			fmt.Println(string(res))
		}
	case WRITE_FILE:
		err := writeToDisk(state, param)
		if err != nil {
			fmt.Println(err)
		}
	default:
		fmt.Println("unknown command")
	}

	return
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
		CheckMin:         3,
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
		execIoCommand(&state)
	}

	return nil
}
