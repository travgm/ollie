package spellcheck

import (
	"strings"
	"testing"
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
	entries := `jupiter
neptune
earth
hello
somehing
random`

	dict := &Dict{MaxSuggest: 3}
	err := dict.load(strings.NewReader(entries))
	if err != nil {
		t.Fatalf("load entries: %v", err)
	}

	var user string = "cello"
	var want = [3]string{"hello", "earth", "random"}
	suggestions, err := dict.Lookup(user)
	if len(suggestions) != 3 {
		t.Fatalf("want %d suggestions, got %v", 3, suggestions)
	}
	var got [3]string
	copy(got[:], suggestions)
	if want != got {
		t.Errorf("CheckWord(%q) = %v, want %v", user, got, want)
	}
}
