package igc

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("sample.igc")
	if err != nil {
		t.Error(err)
	}
	if _, err := NewFlight(f); err != nil {
		t.Error(err)
	}
}
