// Package igc represents the IGC format as described in http://carrier.csi.cam.ac.uk/forsterlewis/soaring/igc_file_format/
package igc

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// Fix represents a B record
type Fix struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	validity  rune
	pressure  int
	gnss      int
}

// NewFix returns a new Fix
func NewFix(date, line string) Fix {
	var err error
	t, err := time.Parse("020106 150405", fmt.Sprintf("%s %s", date, line[1:7]))
	if err != nil {
		log.Fatal(err)
	}
	p, _ := strconv.Atoi(line[25:30])
	g, _ := strconv.Atoi(line[30:35])
	return Fix{
		Time:      t,
		Latitude:  ParseLatitude(line[7:15]),
		Longitude: ParseLongitude(line[15:24]),
		validity:  rune(line[24]),
		pressure:  p,
		gnss:      g,
	}
}

// Coord returns the coordinates
func (f Fix) Coord() string {
	if f.Latitude > 0 && f.Longitude > 0 {
		return fmt.Sprintf("%f,%f", f.Latitude, f.Longitude)
	}
	return ""
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
		switch v.validity {
		case 'A':
			if v.gnss > 0 {
				return p[i]
			}
		case 'V':
			return p[i]
		default:
			log.Fatal("invalid fix")
		}
	}
	return Fix{}
}

// Landing returns the landing fix
func (p FixSlice) Landing() Fix {
	return p[len(p)-1]
}
