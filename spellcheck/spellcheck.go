// MIT License
//
// # Copyright (c) 2024 Travis Montoya
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

// Spellchecking routines for the spellchecker called from editor.go. This is a
// very mediocre implementation of the levenshtein distance algorithm converted
// from psuedo code.
package spellcheck

import (
	"bufio"
	"io"
	"math"
	"os"
	"slices"
	"strings"
	"errors"
)

// Dict holds information regarding the dictionary we have loaded and some
// behavioral settings such as MaxSuggest which limits the amount of suggestions
// we can receive back from CheckWord
//
// MaxSuggest is the maximum suggestions to return for each word spelled wrong
type Dict struct {
	dictionary []string
	MaxSuggest int
}

func NewSpellchecker(dictionaryPath string, suggestions int) (*Dict, error) {
	d := &Dict{MaxSuggest: suggestions}
	err := d.LoadFromFile(dictionaryPath)
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return d, nil
}

func (d *Dict) LoadFromFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.LoadWordlist(f)
}

func (d *Dict) LoadWordlist(r io.Reader) error {
	ws := bufio.NewScanner(r)
	for ws.Scan() {
		d.dictionary = append(d.dictionary, ws.Text())
	}

	if err := ws.Err(); err != nil {
		return err
	}

	return nil
}

func (d *Dict) CheckWord(word string) ([]string, error) {
	wordChoices := make([]string, 1)

	if word == "" || word == " " || slices.Contains(d.dictionary, word) {
		return nil, nil
	}

	// This needs to be refactored as this will have horrible runtime as we iterate the
	// dictionary d.MaxSuggest amount of times to find the best words to fill the
	// suggestions
	for _ = range d.MaxSuggest {
		bestWord := ""
		bestDistance := math.MaxFloat64
		for _, w := range d.dictionary {
			if slices.Contains(wordChoices, w) {
				continue
			}
			dist := LevDistance(word, w)
			if dist < bestDistance {
				bestDistance = dist
				bestWord = w
			}
		}

		if slices.Contains(wordChoices, bestWord) == false && bestWord != "" {
			wordChoices = append(wordChoices, strings.TrimSpace(bestWord))

		}

	}

	return wordChoices, nil

}

// A mediocre Levenshtein Distance algorithm
// https://en.wikipedia.org/wiki/Levenshtein_distance
func LevDistance(word string, dictWord string) float64 {
	// lm[i][j] holds the distance between [i] chars of word and [j] chars of dictWord
	lm := make([][]float64, len(word)+1)
	for i := range lm {
		lm[i] = make([]float64, len(dictWord)+1)
	}

	// Init our 2D slice with increasing values 0..len(word) and 0..len(dictWord)
	// This is used to set the base case and calculate the minimum edits for the word
	for i := 0; i <= len(word); i++ {
		lm[i][0] = float64(i)
	}

	for j := 0; j <= len(dictWord); j++ {
		lm[0][j] = float64(j)
	}

	// Calculate edits for each character of our word and dictionary word
	// return the minimum edits required to change the word into the other word
	for i := 1; i <= len(word); i++ {
		for j := 1; j <= len(dictWord); j++ {
			cost := 0
			if word[i-1] != dictWord[j-1] {
				cost = 1
			}

			// Three possible options:
			// current row + 1 = deletion
			// current col + 1 = insertion
			// prev value in diag + 1 if characters are different or + 0 if same
			lm[i][j] = math.Min(lm[i-1][j]+1,
				math.Min(lm[i][j-1]+1, lm[i-1][j-1]+float64(cost)))
		}
	}
	return lm[len(word)][len(dictWord)]
}

