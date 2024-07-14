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

// Package spellcheck implements a dictionary providing suggestions for incorrectly spelled words.
package spellcheck

import (
	"bufio"
	"io"
	"math"
	"os"
	"slices"
	"strings"
)

// Dict represents a dictionary.
type Dict struct {
	// MaxSuggest is the maximum number of suggestions to return
	// for each word spelled incorrectly.
	MaxSuggest int
	dictionary []string
}

// NewDict returns a new Dict with entries read from the named file.
// The format of the entries is a newline-delimited text file.
// Each line contains a single correctly spelled word.
func NewDict(name string) (*Dict, error) {
	d := &Dict{}
	return d, d.loadFromFile(name)
}

func (d *Dict) loadFromFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.load(f)
}

func (d *Dict) load(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		d.dictionary = append(d.dictionary, sc.Text())
	}
	return sc.Err()
}

// Lookup returns a list of sugggested spellings (up to d.MaxSuggest) for the given word.
// An empty slice is returned if word is present in the dictionary.
func (d *Dict) Lookup(word string) ([]string, error) {
	var wordChoices []string

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

		if slices.Contains(wordChoices, bestWord) == false {
			wordChoices = append(wordChoices, strings.TrimSpace(bestWord))

		}

	}

	return wordChoices, nil

}

// LevDistance returns an estimated "cost" to change word to dictWord.
// It uses a mediocre implementation of the [Levenshtein distance] algorithm.
//
// [Levenshtein distance]: https://en.wikipedia.org/wiki/Levenshtein_distance
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
