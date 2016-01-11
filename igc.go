package igc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	//
	GoogleMapsApiUrl = "https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f"
)

var (
	takeoff, landing Finder
	MaxDistance      = 300
	ErrWrongSuffix   = errors.New("wrong suffix")
)

type Finder interface {
	Find(float64, float64) (string, int)
}

func RegisterTakeoffSiteSource(f Finder) {
	takeoff = f
}

func RegisterLandingSiteSource(f Finder) {
	landing = f
}

type Fix struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	Validity  rune
	Pressure  int
	GNSS      int
}

type FixSlice []Fix

func (p FixSlice) Len() int {
	return len(p)
}

func (p FixSlice) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}

func (p FixSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

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

func (p FixSlice) Landing() Fix {
	return p[len(p)-1]
}

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
}

func NewFlight(r io.Reader) (*Flight, error) {
	flight := &Flight{}
	flight.Fixes = FixSlice{}
	if err := flight.parse(r); err != nil {
		return nil, err
	}
	if err := flight.evaluate(); err != nil {
		return nil, err
	}
	return flight, nil
}

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
		f.Site = strings.Split(line, ": ")[1]
	}
}

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

func (f *Flight) Record() []string {
	return []string{f.Date.Format("02.01.2006"), f.TakeOff.Format("15:04"), f.TakeOffSite, f.Landing.Format("15:04"), f.LandingSite, fmt.Sprintf("%.2f", f.Duration.Minutes())}
}

type Flights []*Flight

func (f Flights) Len() int {
	return len(f)
}

func (f Flights) Less(i, j int) bool {
	return f[i].TakeOff.Before(f[j].TakeOff)
}

func (f Flights) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func ParseLatitude(l string) float64 {
	deg, err := strconv.Atoi(l[0:2])
	if err != nil {
		log.Fatal(err)
	}
	min, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", l[2:4], l[4:7]), 32)
	if err != nil {
		log.Fatal(err)
	}
	switch l[7] {
	case 'N':
		return float64(deg) + float64(min/60)
	case 'S':
		return float64(-deg) + float64(min/60)
	default:
		log.Fatal(ErrWrongSuffix)
	}
	return 0.0 // never reached
}

func ParseLongitude(l string) float64 {
	deg, err := strconv.Atoi(l[0:3])
	if err != nil {
		log.Fatal(err)
	}
	min, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", l[3:5], l[5:8]), 64)
	if err != nil {
		log.Fatal(err)
	}
	switch l[8] {
	case 'E':
		return float64(deg) + min/60
	case 'W':
		return float64(-deg) + min/60
	default:
		log.Fatal(ErrWrongSuffix)
	}
	return 0.0 // never reached
}

func LookupTakeOffSite(lat, lon float64) string {
	var place string
	coord := fmt.Sprintf("(%.4f, %.4f)", lat, lon)
	if takeoff != nil {
		if place, dist := takeoff.Find(lat, lon); len(place) > 0 && dist <= MaxDistance {
			return fmt.Sprintf("%s %s", place, coord)
		}
	}
	place = LookupPlaceWithGoogleMaps(lat, lon)
	if len(place) > 0 {
		return fmt.Sprintf("%s %s", place, coord)
	}
	return coord
}

func LookupLandingSite(lat, lon float64) string {
	var place string
	coord := fmt.Sprintf("(%.4f, %.4f)", lat, lon)
	if landing != nil {
		if place, dist := landing.Find(lat, lon); len(place) > 0 && dist <= MaxDistance {
			return fmt.Sprintf("%s %s", place, coord)
		}
	}
	if len(place) > 0 {
		return fmt.Sprintf("%s %s", place, coord)
	}
	place = LookupPlaceWithGoogleMaps(lat, lon)
	if len(place) > 0 {
		return fmt.Sprintf("%s %s", place, coord)
	}
	return coord
}

func LookupPlaceWithGoogleMaps(lat, lon float64) string {
	var result struct {
		Results []struct {
			Address string `json:"formatted_address"`
		} `json:"results"`
		Status string `json:"status"`
	}
	resp, err := http.Get(fmt.Sprintf(GoogleMapsApiUrl, lat, lon))
	time.Sleep(75 * time.Millisecond) // kind of throttling
	if err != nil {
		log.Println(err)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		return ""
	}
	if result.Status != "OK" {
		return ""
	}
	return result.Results[0].Address
}
