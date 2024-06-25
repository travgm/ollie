package spellcheck

import (
	"bufio"
	"os"
	"math"
	"slices"
)

// Dict holds information regarding the dictionary we have loaded and some
// behavioral settings such as MaxSuggest which limits the amount of suggestions
// we can receive back from CheckWord
type Dict struct {
	WordFile string
	dictionary []string
	MaxSuggest int
}

func (d* Dict) LoadWordlist(words string) error {
	w, err := os.Open(words)
	if err != nil {
		return err
	}
	defer w.Close()

	ws := bufio.NewScanner(w)
	for ws.Scan() {
		d.dictionary = append(d.dictionary, ws.Text())
	}

	if err := ws.Err(); err != nil {
		return err
	}

	return nil
}

func (d* Dict) CheckWord(word string) ([]string, error) {
	wordChoices := make([]string, d.MaxSuggest)
	bestDistance := math.MaxFloat64
	var bestWord string

	// This needs to be refactored as this will have horrible runtime as we iterate the
	// dictionary d.MaxSuggest amount of times to find the best words to fill the
	// suggestions
	for _ = range d.MaxSuggest {
		for _, w := range d.dictionary {
			dist := levDistance(word, w)
			if dist < bestDistance {
				bestDistance = dist
				bestWord = w
			}
		}

		if slices.Contains(wordChoices, bestWord) == false {
			wordChoices = append(wordChoices, bestWord)
		}

	}

	return wordChoices, nil

}

// A mediocre Levenshtein Distance algorithm
func levDistance(word string, dictWord string) float64 {
	// Create 2D array to hold word as the horizontal and dictWord as the vertical
	// lm[i][j] holds the distance between [i] chars of word and [j] chars of dictWord
	lm := make([][]float64, len(word))
	for i := range lm {
		lm[i] = make([]float64, len(dictWord))
	}
	
	// Start with an initial cost of 0
	cost := 0
	for i := range word {
		for j := range dictWord {
			// Simple character comparison, this will be used to determine the edits
			// needed to correct the word (del, insert, sub) 
			// if cost == 0 no edit needed as it would be lm[i-1][j-1] + 0
			if word[i-1] == dictWord[j-1] {
				cost = 0
			} else {
				cost = 1
			}

			lm[i][j] = math.Min(lm[i-1][j] + 1, 
				math.Min(lm[i][j-1] + 1, lm[i-1][j-1] + float64(cost)))
		}
	}
	return lm[len(word)][len(dictWord)]

}


