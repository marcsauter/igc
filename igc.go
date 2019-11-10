// Package igc represents the IGC format as described in http://carrier.csi.cam.ac.uk/forsterlewis/soaring/igc_file_format/
package igc

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

const (
	// GoogleMapsAPIURL default URL for Google Maps API
	GoogleMapsAPIURL = "https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f"
)

var (
	// MaxDistance is the maximal distance to the nearest known location
	MaxDistance = 300
	// ErrWrongSuffix latitude with other than N/S suffix or longitude with other then E/W suffix
	ErrWrongSuffix = errors.New("wrong suffix")
)

// Locator is the interface that wraps the basic Lookup method
//
// Lookup finds the nearest location to the given latitude and
// longitude and returns the location name and the distance to
// this location.
type Locator interface {
	Lookup(latitude, longitude float64) (location string, distance int, err error)
}

// ParseLatitude ...
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

// ParseLongitude ...
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
