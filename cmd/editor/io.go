package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"git.sr.ht/~travgm/ollie/search"
)

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

// Save words into state line buffer
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