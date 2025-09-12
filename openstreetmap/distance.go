// Package openstreetmap provides minimal geocoding and distance utilities
// using OpenStreetMap's Nominatim service and the Haversine formula.
package openstreetmap

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

// GeocodeResult maps the minimal fields we need from Nominatim.
// The Nominatim "search" endpoint returns an array; we need only the first
// element's "lat" and "lon" fields. They are strings in the payload and must
// be parsed to float64. Keeping this type minimal reduces coupling to the API.
type geocodeResult []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// geocode queries Nominatim for a single result and returns lat, lon.
//   - Builds a request to https://nominatim.openstreetmap.org/search with:
//     q=<query>, format=jsonv2, limit=1, addressdetails=0, countrycodes=at
//     This restricts results to Austria and asks for the newest JSON format.
//   - Applies a context (deadline/timeout) from the caller.
//   - Sets a custom User-Agent per Nominatim usage policy. Requests without a
//     valid UA may be throttled or rejected.
//   - Exits the process on network/HTTP/JSON errors via log.Fatal.
//   - Returns the first match's latitude and longitude as float64.
func geocode(ctx context.Context, query string) (float64, float64) {
	// Build request URL for the "search" endpoint.
	u, _ := url.Parse("https://nominatim.openstreetmap.org/search")
	q := u.Query()
	q.Set("q", query)            // Free-text query, e.g., "4020 Linz, Austria"
	q.Set("format", "jsonv2")    // Use the v2 JSON schema
	q.Set("limit", "1")          // Request at most one result
	q.Set("addressdetails", "0") // Exclude verbose address breakdown to save bytes
	q.Set("countrycodes", "at")  // Restrict search to Austria (ISO 3166-1 alpha-2)
	u.RawQuery = q.Encode()

	// Prepare HTTP request with context. Context ensures the request will be
	// cancelled when the deadline is exceeded or the caller cancels.
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	// Set a descriptive User-Agent as required by Nominatim policy.
	// Replace contact@example.com with a real contact address for production use.
	req.Header.Set("User-Agent", "plz-distance-tool/1.0 (contact@example.com)")

	// Execute the request using the default client. For production, consider a
	// custom http.Client with tuned Transport (connection pooling, timeouts).
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Network error, DNS failure, context timeout, etc. Fail fast.
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Validate successful HTTP status. Nominatim may return 429 (Too Many Requests)
	// when rate limited, or 5xx on server error. We fail fast here.
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("geocoding failed: %s", resp.Status)
	}

	// Decode JSON array response into our minimal struct.
	var res geocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatal(err)
	}
	if len(res) == 0 {
		// No candidates found for the query; fail fast to surface the issue.
		log.Fatalf("no geocoding result for %q", query)
	}

	// Parse coordinate strings to float64. If parsing fails, the payload was invalid
	// or unexpected; fail fast to avoid propagating bad data downstream.
	lat, err := strconv.ParseFloat(res[0].Lat, 64)
	if err != nil {
		log.Fatal(err)
	}
	lon, err := strconv.ParseFloat(res[0].Lon, 64)
	if err != nil {
		log.Fatal(err)
	}
	return lat, lon
}

// haversineKM computes great-circle distance in kilometers between two WGS84
// coordinates using the Haversine formula.
//
// Inputs:
//   - lat1, lon1: Latitude and longitude of point A in decimal degrees.
//   - lat2, lon2: Latitude and longitude of point B in decimal degrees.
//
// Method:
//   - Convert degrees to radians.
//   - Apply the Haversine formula to compute central angle c.
//   - Multiply by mean Earth radius R to obtain distance.
//
// Accuracy:
//   - Uses a fixed mean Earth radius R = 6371.0 km. This is suitable for
//     many applications. If high precision is required, consider using a
//     more accurate ellipsoidal model (e.g., Vincenty or Karney).
func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	φ1, λ1 := toRad(lat1), toRad(lon1)
	φ2, λ2 := toRad(lat2), toRad(lon2)

	// Differences in radians
	dφ := φ2 - φ1
	dλ := λ2 - λ1

	// Haversine formula
	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Mean Earth radius in kilometers. For higher accuracy, adjust by latitude.
	const R = 6371.0
	return R * c
}

// Distance returns the integer distance in kilometers between an input Austrian
// postal code (ZIP) and the fixed destination "4020 Linz, Austria".
func Distance(zip string) int {
	// Build source query by pairing the ZIP with the country for better disambiguation.
	src := fmt.Sprintf("%s, Austria", zip)

	// Fixed destination for this tool: postal district 4020 in Linz, Austria.
	dst := "4020 Linz, Austria"

	// Bound the lifetime of both HTTP requests to Nominatim.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Geocode source and destination to WGS84 coordinates.
	lat1, lon1 := geocode(ctx, src)
	lat2, lon2 := geocode(ctx, dst)

	// Compute great-circle distance between the two points.
	km := haversineKM(lat1, lon1, lat2, lon2)

	// Return the truncated integer number of kilometers.
	return int(km)
}
