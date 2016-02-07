// Package igc represents the IGC format as described in http://carrier.csi.cam.ac.uk/forsterlewis/soaring/igc_file_format/
package igc

import (
	"log"
	"time"
)

// Fix represents a B record
type Fix struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	Validity  rune
	Pressure  int
	GNSS      int
}

// FixSlice represents a slice of B records
type FixSlice []Fix

// Len for sort interface
func (p FixSlice) Len() int {
	return len(p)
}

// Less for sort interface
func (p FixSlice) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}

// Swap for sort interface
func (p FixSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// TakeOff returns the takeoff fix
func (p FixSlice) TakeOff() Fix {
	for i, v := range p {
		switch v.Validity {
		case 'A':
			if v.GNSS > 0 {
				return p[i]
			}
		case 'V':
			return p[i]
		default:
			log.Fatal()
		}
	}
	return p[0]
}

// Landing returns the landing fix
func (p FixSlice) Landing() Fix {
	return p[len(p)-1]
}
