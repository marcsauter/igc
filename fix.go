package igc

import (
	"log"
	"time"
)

//
type Fix struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	Validity  rune
	Pressure  int
	GNSS      int
}

//
type FixSlice []Fix

//
func (p FixSlice) Len() int {
	return len(p)
}

//
func (p FixSlice) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}

//
func (p FixSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

//
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

//
func (p FixSlice) Landing() Fix {
	return p[len(p)-1]
}
