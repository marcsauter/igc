package igc

import (
	"bufio"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// Flight represents a recorded flight
type Flight struct {
	RawDate     string
	Date        time.Time
	Glider      string
	TakeOff     Fix
	Site        string
	TakeOffSite string
	Landing     Fix
	LandingSite string
	Duration    time.Duration
	Fixes       FixSlice
	Filename    string
	Comment     string
}

// NewFlightFromFile reaturns a flight evaluated from a igc file
func NewFlightFromFile(path string) (*Flight, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	flight := &Flight{}
	flight.Filename = path
	flight.Fixes = FixSlice{}
	if err := flight.parse(f); err != nil {
		return nil, err
	}
	if err := flight.evaluate(); err != nil {
		return nil, err
	}
	return flight, nil
}

// parse a flight
func (f *Flight) parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// ignore blank lines
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'H':
			f.parseHrecord(line)
		case 'B':
			f.Fixes = append(f.Fixes, NewFix(f.RawDate, line))
		default:
		}
	}
	return nil
}

// parseHrecord
func (f *Flight) parseHrecord(line string) {
	switch line[2:5] {
	case "DTE":
		var err error
		f.RawDate = convert(line[5:11])
		f.Date, err = time.Parse("020106", f.RawDate)
		if err != nil {
			log.Fatal(err)
		}
	case "SIT":
		if d := strings.Split(line, ": "); len(d) == 2 {
			f.Site = convert(d[1])
		}
	case "GTY":
		if d := strings.Split(line, ": "); len(d) == 2 {
			f.Glider = convert(d[1])
		}
	}
}

// evaluate
func (f *Flight) evaluate() error {
	sort.Sort(f.Fixes)
	f.TakeOff = f.Fixes.TakeOff()
	f.TakeOffSite = LookupTakeOffSite(f.TakeOff.Latitude, f.TakeOff.Longitude)
	f.Landing = f.Fixes.Landing()
	f.LandingSite = LookupLandingSite(f.Landing.Latitude, f.Landing.Longitude)
	f.Duration = f.Landing.Time.Sub(f.TakeOff.Time)
	return nil
}

// Flights
type Flights []*Flight

// NewFlights
func NewFlights() *Flights {
	return &Flights{}
}

// Add
func (f *Flights) Add(flight *Flight) error {
	*f = append(*f, flight)
	return nil
}

// Len
func (f Flights) Len() int {
	return len(f)
}

// Less
func (f Flights) Less(i, j int) bool {
	return f[i].TakeOff.Time.Before(f[j].TakeOff.Time)
}

// Swap
func (f Flights) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// convert
func convert(s string) string {
	s = strings.Trim(s, " ")
	r := make([]rune, len(s))
	for i := range s {
		r[i] = rune(s[i])
	}
	return string(r)
}
