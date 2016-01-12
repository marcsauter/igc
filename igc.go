package igc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	//
	GoogleMapsApiUrl = "https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f"
)

var (
	takeoff, landing Finder
	//
	MaxDistance = 300
	//
	ErrWrongSuffix = errors.New("wrong suffix")
)

//
type Finder interface {
	Find(float64, float64) (string, int)
}

//
func RegisterTakeoffSiteSource(f Finder) {
	takeoff = f
}

//
func RegisterLandingSiteSource(f Finder) {
	landing = f
}

//
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

//
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

//
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

//
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

//
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
