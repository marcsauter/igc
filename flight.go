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

//
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

//
func NewFlight(path string) (*Flight, error) {
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

//
func (f *Flight) parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
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

//
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
		f.Site = convert(strings.Split(line, ": ")[1])
	case "GTY":
		f.Glider = convert(strings.Split(line, ": ")[1])
	}
}

//
func (f *Flight) evaluate() error {
	sort.Sort(f.Fixes)
	f.TakeOff = f.Fixes.TakeOff()
	f.TakeOffSite = LookupTakeOffSite(f.TakeOff.Latitude, f.TakeOff.Longitude)
	f.Landing = f.Fixes.Landing()
	f.LandingSite = LookupLandingSite(f.Landing.Latitude, f.Landing.Longitude)
	f.Duration = f.Landing.Time.Sub(f.TakeOff.Time)
	return nil
}

//
type Flights []*Flight

//
func NewFlights() *Flights {
	return &Flights{}
}

//
func (f *Flights) Add(flight *Flight) error {
	*f = append(*f, flight)
	return nil
}

//
func (f Flights) Len() int {
	return len(f)
}

//
func (f Flights) Less(i, j int) bool {
	return f[i].TakeOff.Time.Before(f[j].TakeOff.Time)
}

//
func (f Flights) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func convert(s string) string {
	s = strings.Trim(s, " ")
	r := make([]rune, len(s))
	for i, _ := range s {
		r[i] = rune(s[i])
	}
	return string(r)
}
