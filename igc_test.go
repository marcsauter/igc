package igc_test

import (
	"testing"

	"github.com/marcsauter/igc"
)

func TestParse(t *testing.T) {
	if _, err := igc.NewFlightFromFile("sample.igc"); err != nil {
		t.Error(err)
	}
}
