// MIT License
//
// # Copyright (c) 2024 Travis Montoya and the ollie contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This is the main editor code for ollie that handles the event loop, processing
// the commands entered and starting the go routine for the spellchecker.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"git.sr.ht/~travgm/ollie/conf"
	"git.sr.ht/~travgm/ollie/olliefile"
	"git.sr.ht/~travgm/ollie/version"
)

// Editor state, this should hold everything we need to be passed around
// and modify state. The only channels we use are a send and receive one
// for spellchecking and a done channel to signify the program is exiting
//
// command should only be one of the valid editor const commands
type State struct {
	channels   Channels
	dictionary string
	command    string
	wordInput  *bufio.Scanner
	ollie      *olliefile.File
	conf       *conf.Settings
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
	HELP          = "h"
	REG_EXP       = "/"
	READ_IN_FILE  = "r"
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
		// Move this to io.go
		params := strings.Split(param, " ")
		var spellFile string

		if len(params) == 2 {
			spellFile = params[1]
		}

		switch params[0] {
		case "on":
			if !state.channels.SpellRunning {
				go execSpellchecker(&state.channels, spellFile)
			} else if state.channels.SpellRunning && !state.channels.ShouldSpellcheck {
				state.channels.ShouldSpellcheck = true
			}
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
	case REG_EXP:
		err := searchWithRegExp(state, param)
		if err != nil {
			fmt.Println(err)
		}
	case READ_IN_FILE:
		err := readInFile(state, param)
		if err != nil {
			fmt.Println(err)
		}
	case HELP:
		showHelpMenu()
	default:
		fmt.Println("unknown command")
	}

	return
}

func initEditor(filename string, spell string) (State, error) {
	of := &olliefile.File{Name: "junk.ollie"}
	if filename != "" {
		of.Name = filename
		of.CreateFile()
	}
	config, err := conf.FromFile(conf.DefaultConfFile)
	if !errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}

	shouldSpell := len(spell) > 0

	spChannels := Channels{
		ShouldSpellcheck: shouldSpell,
		CheckMin:         3,
		Spelling:         make(chan []string, 1),
		Spellres:         make(chan []string, 1),
		Done:             make(chan string, 1),
	}

	state := State{
		channels:   spChannels,
		dictionary: spell,
		wordInput:  bufio.NewScanner(os.Stdin),
		ollie:      of,
		conf:       config,
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
	spellFlag := flag.String("spcheck", "", "Turn spellchecking on with provided dictionary, default is off")

	flag.Parse()
	if *aboutFlag {
		version.DisplayAbout()
		return nil
	}

	state, err := initEditor(flag.Arg(0), *spellFlag)
	if err != nil {
		return err
	}

	if state.channels.ShouldSpellcheck {
		go execSpellchecker(&state.channels, *spellFlag)
	}

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
