package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type AppRating struct {
	AppID       string  `json:"app_id"`
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"rating_count"`
}

func (r AppRating) CSVHeaders() []string {
	return []string{"app_id", "name", "rating", "rating_count"}
}

func (r AppRating) CSVRow() []string {
	return []string{r.AppID, r.Name, fmt.Sprintf("%.2f", r.Rating), intToStr(r.RatingCount)}
}

// LookupBaseURL is the base URL for iTunes Lookup API. Tests can override this.
var LookupBaseURL = "https://itunes.apple.com"

var lookupHTTPClient = &http.Client{Timeout: 30 * time.Second}

var (
	appIDPattern   = regexp.MustCompile(`^\d+$`)
	countryPattern = regexp.MustCompile(`^[a-zA-Z]{2}$`)
)

type lookupResponse struct {
	ResultCount int            `json:"resultCount"`
	Results     []lookupResult `json:"results"`
}

type lookupResult struct {
	TrackID            int     `json:"trackId"`
	TrackName          string  `json:"trackName"`
	AverageUserRating  float64 `json:"averageUserRating"`
	UserRatingCount    int     `json:"userRatingCount"`
}

func LookupAppRating(ctx context.Context, appID, country string) (AppRating, error) {
	if !appIDPattern.MatchString(appID) {
		return AppRating{}, fmt.Errorf("invalid app ID %q: must be numeric", appID)
	}
	if !countryPattern.MatchString(country) {
		return AppRating{}, fmt.Errorf("invalid country code %q: must be a 2-letter code", country)
	}

	url := fmt.Sprintf("%s/lookup?id=%s&country=%s", LookupBaseURL, appID, country)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AppRating{}, fmt.Errorf("creating lookup request: %w", err)
	}

	resp, err := lookupHTTPClient.Do(req)
	if err != nil {
		return AppRating{}, fmt.Errorf("iTunes lookup: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return AppRating{}, fmt.Errorf("iTunes lookup error %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AppRating{}, fmt.Errorf("reading lookup response: %w", err)
	}

	var result lookupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return AppRating{}, fmt.Errorf("parsing lookup response: %w", err)
	}

	if len(result.Results) == 0 {
		return AppRating{}, nil
	}

	r := result.Results[0]
	return AppRating{
		AppID:       appID,
		Name:        r.TrackName,
		Rating:      r.AverageUserRating,
		RatingCount: r.UserRatingCount,
	}, nil
}

func LookupAppRatings(ctx context.Context, appIDs []string, country string, warn func(string, error)) map[string]AppRating {
	ratings := make(map[string]AppRating, len(appIDs))
	for _, id := range appIDs {
		rating, err := LookupAppRating(ctx, id, country)
		if err != nil {
			if warn != nil {
				warn(id, err)
			}
			continue
		}
		ratings[id] = rating
	}
	return ratings
}
