package speedtest

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
)

// coloNames maps IATA airport codes to human-readable city names.
var coloNames = map[string]string{
	"ATL": "Atlanta, GA",
	"IAD": "Ashburn, VA",
	"BOS": "Boston, MA",
	"BUF": "Buffalo, NY",
	"CLT": "Charlotte, NC",
	"ORD": "Chicago, IL",
	"CMH": "Columbus, OH",
	"DFW": "Dallas, TX",
	"DEN": "Denver, CO",
	"DTW": "Detroit, MI",
	"HNL": "Honolulu, HI",
	"IAH": "Houston, TX",
	"IND": "Indianapolis, IN",
	"JAX": "Jacksonville, FL",
	"MCI": "Kansas City, MO",
	"LAS": "Las Vegas, NV",
	"LAX": "Los Angeles, CA",
	"MEM": "Memphis, TN",
	"MIA": "Miami, FL",
	"MSP": "Minneapolis, MN",
	"BNA": "Nashville, TN",
	"EWR": "Newark, NJ",
	"MSY": "New Orleans, LA",
	"JFK": "New York, NY",
	"OMA": "Omaha, NE",
	"PHL": "Philadelphia, PA",
	"PHX": "Phoenix, AZ",
	"PIT": "Pittsburgh, PA",
	"PDX": "Portland, OR",
	"RDU": "Raleigh, NC",
	"SMF": "Sacramento, CA",
	"SLC": "Salt Lake City, UT",
	"SAT": "San Antonio, TX",
	"SAN": "San Diego, CA",
	"SFO": "San Francisco, CA",
	"SJC": "San Jose, CA",
	"SEA": "Seattle, WA",
	"STL": "St. Louis, MO",
	"TPA": "Tampa, FL",
	"YYZ": "Toronto, ON",
	"YVR": "Vancouver, BC",
	"YUL": "Montreal, QC",
	"LHR": "London, UK",
	"CDG": "Paris, FR",
	"FRA": "Frankfurt, DE",
	"AMS": "Amsterdam, NL",
	"NRT": "Tokyo, JP",
	"SIN": "Singapore, SG",
	"SYD": "Sydney, AU",
	"GRU": "São Paulo, BR",
	"ICN": "Seoul, KR",
	"HKG": "Hong Kong, HK",
	"BOM": "Mumbai, IN",
}

// FetchMeta retrieves server metadata from Cloudflare's trace endpoint.
func FetchMeta(ctx context.Context, client *http.Client) (*ServerInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/cdn-cgi/trace", nil)
	if err != nil {
		return nil, fmt.Errorf("creating meta request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching meta: %w", err)
	}
	defer resp.Body.Close()

	info := &ServerInfo{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "ip":
			info.IP = val
		case "colo":
			info.Colo = val
		case "loc":
			info.Location = val
		}
	}

	if city, ok := coloNames[info.Colo]; ok {
		info.ColoCity = city
	} else {
		info.ColoCity = info.Colo
	}

	return info, scanner.Err()
}
