package spellcheck_test

import (
	"testing"

	"git.sr.ht/~travgm/ollie/spellcheck"
)

func TestLevDistance(t *testing.T) {
	userWord := "hello"
	dictWord := "cello"

	dist := spellcheck.LevDistance(userWord, dictWord)
	if dist != 1 {
		t.Fatal("Wrong distance for hello and cello")
	}

}
