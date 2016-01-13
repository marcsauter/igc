package igc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//
type Flight struct {
	RawDate     string
	Date        time.Time
	TakeOff     time.Time
	Site        string
	TakeOffSite string
	Landing     time.Time
	LandingSite string
	Duration    time.Duration
	Fixes       FixSlice
	Filename    string
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
			f.parseBrecord(line)
		default:
		}
	}
	return nil
}

//
func (f *Flight) parseHrecord(line string) {
	switch line[2:5] {
	case "DTE":
		date, err := time.Parse("020106", line[5:11])
		if err != nil {
			log.Fatal(err)
		}
		f.Date = date
		f.RawDate = line[5:11]
	case "SIT":
		sit := strings.Split(line, ": ")[1]
		buf := make([]rune, len(sit))
		for i, b := range sit {
			buf[i] = rune(b)
		}
		f.Site = string(buf)
	}
}

//
func (f *Flight) parseBrecord(line string) {
	var err error
	p := Fix{}
	p.Time, err = time.Parse("020106 150405", fmt.Sprintf("%s %s", f.RawDate, line[1:7]))
	if err != nil {
		log.Fatal(err)
	}
	p.Latitude = ParseLatitude(line[7:15])
	p.Longitude = ParseLongitude(line[15:24])
	p.Validity = rune(line[24])
	p.Pressure, _ = strconv.Atoi(line[25:30])
	p.GNSS, _ = strconv.Atoi(line[30:35])
	f.Fixes = append(f.Fixes, p)
}

//
func (f *Flight) evaluate() error {
	sort.Sort(f.Fixes)
	takeOff := f.Fixes.TakeOff()
	f.TakeOff = takeOff.Time
	f.TakeOffSite = f.Site
	if len(f.TakeOffSite) == 0 {
		f.TakeOffSite = LookupTakeOffSite(takeOff.Latitude, takeOff.Longitude)
	}
	landing := f.Fixes.Landing()
	f.Landing = landing.Time
	f.LandingSite = LookupLandingSite(landing.Latitude, landing.Longitude)
	f.Duration = f.Landing.Sub(f.TakeOff)
	return nil
}

//
func (f *Flight) Record() []string {
	return []string{f.Date.Format("02.01.2006"), f.TakeOff.Format("15:04"), f.TakeOffSite, f.Landing.Format("15:04"), f.LandingSite, fmt.Sprintf("%.2f", f.Duration.Minutes()), f.Filename}
}

//
type Flights []*Flight

//
func (f Flights) Len() int {
	return len(f)
}

//
func (f Flights) Less(i, j int) bool {
	return f[i].TakeOff.Before(f[j].TakeOff)
}

//
func (f Flights) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//
func (f *Flights) Output() *[][]string {
	s := [][]string{}
	for _, flight := range *f {
		s = append(s, flight.Record())
	}
	return &s
}
