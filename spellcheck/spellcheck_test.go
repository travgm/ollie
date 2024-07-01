package spellcheck

import "testing"

func TestLevDistance(t *testing.T) {
	user := "hello"
	dict := "cello"

	dist := LevDistance(user, dict)
	var want float64 = 1
	if dist != want {
		t.Fatalf("LevDistance(%q, %q) = %f, want %f", user, dict, dist, want)
	}
}
