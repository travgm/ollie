package spellcheck

import (
	"bufio"
	"math"
	"os"
	"slices"
	"strings"
)

// Dict holds information regarding the dictionary we have loaded and some
// behavioral settings such as MaxSuggest which limits the amount of suggestions
// we can receive back from CheckWord
type Dict struct {
	WordFile   string
	dictionary []string
	MaxSuggest int
}

func (d *Dict) LoadWordlist() error {
	w, err := os.Open(d.WordFile)
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

func (d *Dict) CheckWord(word string) ([]string, error) {
	wordChoices := make([]string, 1)
	bestDistance := math.MaxFloat64
	var bestWord string

	// This needs to be refactored as this will have horrible runtime as we iterate the
	// dictionary d.MaxSuggest amount of times to find the best words to fill the
	// suggestions
	for _ = range d.MaxSuggest {
		for _, w := range d.dictionary {
			if strings.TrimSpace(w) == "" {
				continue
			}
			dist := levDistance(word, w)
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

// A mediocre Levenshtein Distance algorithm
// https://en.wikipedia.org/wiki/Levenshtein_distance
func levDistance(word string, dictWord string) float64 {
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
