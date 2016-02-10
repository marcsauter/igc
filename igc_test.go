package igc

import "testing"

func TestParse(t *testing.T) {
	if _, err := NewFlight("sample.igc"); err != nil {
		t.Error(err)
	}
}
