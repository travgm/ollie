package spellcheck

import (
	"testing"
	"strings"
)

func TestLevDistance(t *testing.T) {
	user := "hello"
	dict := "cello"

	dist := LevDistance(user, dict)
	var want float64 = 1
	if dist != want {
		t.Fatalf("LevDistance(%q, %q) = %f, want %f", user, dict, dist, want)
	}
}

func TestGetWords(t *testing.T) {
	testDictionary := "jupiter\nneptune\nearth\nhello\nsomehing\nrandom"
	d, err := NewSpellchecker("", 3)
	if err != nil {
		t.Fatalf("NewSpellchecker: %v", err)
	}
	r := strings.NewReader(testDictionary)
	
	err = d.LoadWordlist(r)
	if err != nil {
		t.Fatalf("LoadWordlist: %v", err)
	}

	var want string = "hello"
	var user string = "cello"
	words, err := d.CheckWord(user)
	if len(words) != 3 && words[0] != want {
		t.Fatalf("CheckWord(%s) = %s, want %s", user, words[0], want)
	}
}


