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
package main

import (
	"fmt"
	"os"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"git.sr.ht/~travgm/ollie/search"
  "git.sr.ht/~travgm/ollie/spellcheck"
)

type Channels struct {
	// This can be turned on and off to disable spellchecking
	ShouldSpellcheck bool

	// This is only set when the go routine is running
	SpellRunning bool

	// The minimum amount of characters a line must have before we send it to the spellchecker
	CheckMin int

	// Spellcheck send/receive channels
	Spelling chan []string
	Spellres chan []string
	Done     chan string
}

// Returns a main command and its parameters.
//
// NOTE: the second string *param* can be a single string of multiple parameters
// depending on what the main command is. It is up to that command to parse the
// params accordingly.
func parseCommandArgs(state *State) (string, string, error) {
	c := strings.Split(state.command, " ")

	cmdLen := len(c)
	cmd := c[0]

	param := ""
	if cmdLen > 1 {
		param = strings.Join(c[1:], " ")
	}

	return cmd, param, nil
}


// Executes a given string to the system shell
func shellCommand(command string) ([]byte, error) {
	if len(command) < 1 {
		return nil, fmt.Errorf("no command specified to run")
	}
	param := strings.Split(command, " ")
	scmd := exec.Command(param[0], param[1:]...)
	res, err := scmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return res, nil

}

// This will delete the last line from the buffer and if there is a valid file
// handle it will delete the last line from the file as well. This later might
// be changed to just delete it from the buffer as the buffer is not written
// to the file until the user types "w" so we could potentially remove a line
// that the user would want in the file.
func deleteLastLine(state *State) (int, error) {
	line := strconv.Itoa(len(state.ollie.Lines))
	if state.ollie.FileHandle != nil {
		err := state.ollie.UpdateLine(line, "")
		if err != nil {
			return -1, err
		}
	}
	state.ollie.Lines = state.ollie.Lines[:len(state.ollie.Lines)-1]
	state.ollie.LineCount -= 1
	return state.ollie.LineCount + 1, nil
}

// Retrieves words from the user and appended to the ollie line buffer. We
// update some of the stats on the file after the user types a line.
func getWords(state *State) error {
	if state == nil {
		return fmt.Errorf("GetWords Error. State is null\n")
	}

	for state.wordInput.Scan() {
		if state.wordInput.Text() == COMMAND_MODE {
			break
		}

		if state.channels.ShouldSpellcheck &&
			len(state.wordInput.Text()) >= state.channels.CheckMin {
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

// Currently utilizing the go stdlib implementation of the boyer-moore string searching algorithm
func searchLinesBuffer(state *State, text string) (bool, error) {
	if text == "" {
		return false, fmt.Errorf("need a text string to search the buffer for")
	}
	found := false
	sf := search.MakeStringFinder(text)
	for i, line := range state.ollie.Lines {
		if sf.Next(line) != -1 {
			fmt.Printf("%d:%s\n", i+1, line)
			found = true
		}
	}
	if found == false {
		fmt.Printf("%s not found in buffer\n", text)
	}
	return found, nil
}

// Writes the ollie line buffer to a specified file, If no file is given it then writes it to
// the ollie.junk file
func writeToDisk(state *State, param string) error {
	if param != "" {
		state.ollie.Name = param
		err := state.ollie.CreateFile()
		if err != nil {
			return err
		}
	} else if state.ollie.FileHandle == nil && state.ollie.Name != "" {
		err := state.ollie.CreateFile()
		if err != nil {
			return err
		}
	}

	bytes, err := state.ollie.WriteFile()
	if err != nil {
		return err
	} else {
		fmt.Printf("Wrote %d bytes to %s\n", bytes, state.ollie.Name)
	}

	return nil
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

// Show the help menu inside the editor so we can see the commands and
// what they do.
func showHelpMenu() {
    fmt.Println("Help Menu:")
    fmt.Printf("%s\t%s\n", WRITE_FILE, "Write file")
    fmt.Printf("%s\t%s\n", APPEND, "Append to file")
    fmt.Printf("%s\t%s\n", FILE_INFO, "File information")
    fmt.Printf("%s\t%s\n", SPELLCHECK, "Turn spellchecking on/off")
    fmt.Printf("%s\t%s\n", FIX_LINE, "Fix line")
    fmt.Printf("%s\t%s\n", EXEC_CMD, "Execute command")
    fmt.Printf("%s\t%s\n", QUIT_EDITOR, "Quit editor")
    fmt.Printf("%s\t%s\n", DEL_LAST_LINE, "Delete last line")
    fmt.Printf("%s\t%s\n", SEARCH_TEXT, "Search text")
    fmt.Printf("%s\t%s\n", COMMAND_MODE, "Command mode")
    fmt.Printf("%s\t%s\n", HELP, "Help") 
}

// Go routine to handle spellchecking
// Dictionary is hardcoded for now until we get config working
func execSpellchecker(channel *Channels, filePath string) {
	dict, err := spellcheck.NewDict(filePath)
	if errors.Is(err, os.ErrNotExist) {
		// If the user supplied dictionary is not found then we try the default
		// location found on most *nix os's
		//
		// We dont worry about it if this cant load because then the spellchecker
		// just returns that there is no suggestion for the word
		fmt.Printf("dictionary file %s not found trying default\n", filePath)
		dict, err = spellcheck.NewDict("/usr/share/dict/words")
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("default dictionary not found. Please specify a dictionary to use spellchecking\n")
			channel.ShouldSpellcheck = false
			channel.SpellRunning = false
			return
		}
	}

	dict.MaxSuggest = 3
	channel.SpellRunning = true

	for {
		select {
		// We received a message to the spellchecker. We spell check the slice
		// and send back a slice that has suggestions.
		case words, ok := <-channel.Spelling:
			if ok {
				var suggestion []string
				for _, word := range words {
					vals, err := dict.Lookup(word)
					if err != nil || len(vals) < 1 {
						continue
					}
					suggestion = append(suggestion, vals...)
				}
				channel.Spellres <- suggestion
			}
		case <-channel.Done:
			return
		}
	}
}
