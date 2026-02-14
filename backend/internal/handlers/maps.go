package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"dsi_interna_sys/internal/utils"
)

type MapsHandler struct{}

type Place struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

func (h *MapsHandler) SearchPlaces(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.RespondBadRequest(w, "Query parameter 'q' is required")
		return
	}

	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		utils.RespondInternalError(w, "Google Maps API key not configured")
		return
	}

	// search using Text Search API (New) or Places API (Legacy)
	// Using Text Search (Legacy) for simplicity: https://maps.googleapis.com/maps/api/place/textsearch/json
	// Actually, let's use the new Text Search (ID-only first then details) or just the old one if enabled?
	// The prompt implies "grab... name, coordinates, address".
	// Let's use the Find Place request or Text Search. Text Search is most robust for "Simpang Lima Semarang".

	endpoint := "https://maps.googleapis.com/maps/api/place/textsearch/json"
	u, _ := url.Parse(endpoint)
	q := u.Query()
	q.Set("query", query)
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		utils.RespondInternalError(w, "Failed to contact Google Maps API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.RespondInternalError(w, "Google Maps API returned error")
		return
	}

	// Parse simplified response
	var googleResp struct {
		Results []struct {
			Name             string `json:"name"`
			FormattedAddress string `json:"formatted_address"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		utils.RespondInternalError(w, "Failed to parse Google Maps response")
		return
	}

	if googleResp.Status != "OK" && googleResp.Status != "ZERO_RESULTS" {
		utils.RespondInternalError(w, "Google Maps API error: "+googleResp.Status)
		return
	}

	places := make([]Place, 0)
	for _, res := range googleResp.Results {
		places = append(places, Place{
			Name:      res.Name,
			Address:   res.FormattedAddress,
			Latitude:  res.Geometry.Location.Lat,
			Longitude: res.Geometry.Location.Lng,
		})
	}

	utils.RespondSuccess(w, "Places found", places)
}
