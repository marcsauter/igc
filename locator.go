package igc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Opt are options for New
type Opt func(s *Site) error

// Site represents a named geographical site
type Site struct {
	r io.Reader
	d int
	u url.URL
}

// New returns a new site instance.
func New(r io.Reader, opts ...Opt) (*Site, error) {
	s := &Site{
		r: r,
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// WithMaxDistance defines the maximal distance to the nearest site
func WithMaxDistance(d int) Opt {
	return func(s *Site) error {
		if d > 0 {
			s.d = d
			return nil
		}
		return errors.New("maximal distance has to be greater then 0")
	}
}

// WithGoogleMapsURL defines the URL
func WithGoogleMapsURL(u url.URL) Opt {
	return func(s *Site) error {
		s.u = u
		return nil
	}
}

// Lookup implements the Locater interface
func (s *Site) Lookup(lat, lon float64) (string, int, error) {
	var place string
	if takeoff != nil {
		if place, dist := takeoff.Find(lat, lon); len(place) > 0 && dist <= MaxDistance {
			return "", 0, place
		}
	}
	place = LookupPlaceWithGoogleMaps(lat, lon)
	if len(place) > 0 {
		return "", 0, place
	}
	return "", 0, ""

	return "", 0, nil
}

//
func (s *Site) lookupGoogleMaps(lat, lon float64) string {
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
