package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidAddress is returned when the geolookup does not find a valid address
	ErrInvalidAddress = errors.New("nothing found for that address")
)

// GetCoordinates calls to the Google Maps API and attempts to find the coordinates for
// a specified address
func GetCoordinates(addr []string) (*GoogleReturn, error) {
	address := URLEncode(strings.Join(addr, " "))
	geoURL := fmt.Sprintf("http://maps.googleapis.com/maps/api/geocode/json?address=%s&sensor=false", address)
	data, err := Fetch(geoURL)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	var geo geoLocation
	json.Unmarshal(data, &geo)
	if geo.Status != "OK" || len(geo.Results) < 1 {
		return nil, ErrInvalidAddress
	}
	ret := &GoogleReturn{
		Lat:              geo.Results[0].Geometry.Location.Lat,
		Long:             geo.Results[0].Geometry.Location.Lng,
		FormattedAddress: geo.Results[0].FormattedAddress,
	}
	return ret, nil
}

// GoogleReturn contains the Lat/Long of the address, as well as the
// Cannonical address
type GoogleReturn struct {
	Lat              float64
	Long             float64
	FormattedAddress string
}

type geoLocation struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		Types []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}
