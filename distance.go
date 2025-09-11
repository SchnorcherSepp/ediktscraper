package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GeocodeResult maps the minimal fields we need from Nominatim
type GeocodeResult []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// geocode queries Nominatim for a single result and returns lat, lon
func geocode(ctx context.Context, query string) (float64, float64, error) {
	// Build request URL
	u, _ := url.Parse("https://nominatim.openstreetmap.org/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("limit", "1")
	q.Set("addressdetails", "0")
	q.Set("countrycodes", "at") // restrict to Austria
	u.RawQuery = q.Encode()

	// Prepare HTTP request with timeout and proper User-Agent per Nominatim policy
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("User-Agent", "plz-distance-tool/1.0 (contact@example.com)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding failed: %s", resp.Status)
	}

	var res GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}
	if len(res) == 0 {
		return 0, 0, fmt.Errorf("no geocoding result for %q", query)
	}

	// Parse coordinates
	lat, err := strconv.ParseFloat(res[0].Lat, 64)
	if err != nil {
		return 0, 0, err
	}
	lon, err := strconv.ParseFloat(res[0].Lon, 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}

// haversineKM computes great-circle distance in kilometers
func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	φ1, λ1 := toRad(lat1), toRad(lon1)
	φ2, λ2 := toRad(lat2), toRad(lon2)

	dφ := φ2 - φ1
	dλ := λ2 - λ1

	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	const R = 6371.0 // mean Earth radius in km
	return R * c
}

func distance(zip string) int {
	src := fmt.Sprintf("%s, Austria", zip)
	dst := "4030 Linz, Austria"

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	lat1, lon1, err := geocode(ctx, src)
	if err != nil {
		log.Fatal(err)
	}
	lat2, lon2, err := geocode(ctx, dst)
	if err != nil {
		log.Fatal(err)
	}

	km := haversineKM(lat1, lon1, lat2, lon2)
	//fmt.Printf("Entfernung %s → %s: %.1f km\n", src, dst, km)
	return int(km)
}
