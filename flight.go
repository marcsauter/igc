package igc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
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
		for i, _ := range sit {
			buf[i] = rune(sit[i])
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
func (f *Flight) Csv(w *csv.Writer) error {
	return w.Write([]string{f.Date.Format("02.01.2006"), f.TakeOff.Format("15:04"), f.TakeOffSite, f.Landing.Format("15:04"), f.LandingSite, fmt.Sprintf("%.2f", f.Duration.Minutes()), f.Filename})
}

//
func (f *Flight) Xlsx(sheet *xlsx.Sheet) error {
	r := sheet.AddRow()
	//
	c1 := r.AddCell()
	c1.SetDate(f.Date)
	c1.NumFmt = "dd.mm.yyyy"
	//
	c2 := r.AddCell()
	c2.SetDateTime(f.TakeOff)
	c2.NumFmt = "hh:mm"
	//
	r.AddCell().SetString(f.TakeOffSite)
	//
	c3 := r.AddCell()
	c3.SetDateTime(f.Landing)
	c3.NumFmt = "hh:mm"
	//
	r.AddCell().SetString(f.LandingSite)
	//
	r.AddCell().SetFloatWithFormat(f.Duration.Minutes(), "0.00")
	//
	if len(f.Comment) > 0 {
		r.AddCell().SetString(f.Comment)
	} else {
		r.AddCell().SetString(f.Filename)
	}
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
	return f[i].TakeOff.Before(f[j].TakeOff)
}

//
func (f Flights) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//
func (f *Flights) Csv(w *csv.Writer) {
	w.Write([]string{"Date", "Takeoff", "Takeoff Site", "Landing", "Landing Site", "Duration", "Filename"})
	for _, flight := range *f {
		flight.Csv(w)
	}
}

//
func (f *Flights) Xlsx(s *xlsx.Sheet) {
	// header
	// 1st line/row
	r0 := s.AddRow()
	ti := r0.AddCell()
	ti.Merge(6, 0) // merge with the following 6 cells
	ti.SetString("Flights")
	// 2nd line/row
	r1 := s.AddRow()
	r1.AddCell().SetString("Date")
	//
	to := r1.AddCell()
	to.Merge(1, 0) // merge with the following cell
	to.SetString("Takeoff")
	r1.AddCell() // cell to merge
	//
	la := r1.AddCell()
	la.Merge(1, 0) // merge with the following cell
	la.SetString("Landing")
	r1.AddCell() // cell to merge
	//
	r1.AddCell().SetString("Duration")
	r1.AddCell().SetString("Filename")
	// 3rd line/row
	r2 := s.AddRow()
	r2.AddCell() // start with an empty cell
	r2.AddCell().SetString("Time")
	r2.AddCell().SetString("Site")
	r2.AddCell().SetString("Time")
	r2.AddCell().SetString("Site")
	// flights
	for _, flight := range *f {
		flight.Xlsx(s)
	}
}
