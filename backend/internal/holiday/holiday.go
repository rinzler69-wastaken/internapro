package holiday

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	apiURL = "https://api-harilibur.vercel.app/api"
)

// Holiday represents a single holiday.
type Holiday struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

// APIHoliday represents the structure of a holiday from the external API.
type APIHoliday struct {
	HolidayDate     string `json:"holiday_date"`
	HolidayName     string `json:"holiday_name"`
	IsNationalHoliday bool   `json:"is_national_holiday"`
}

// GetHolidays fetches Indonesian national holidays for a given year.
// It uses an external API and has a fallback to a hardcoded list.
func GetHolidays(year int) ([]Holiday, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return getFallbackHolidays(year), fmt.Errorf("failed to get holidays from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getFallbackHolidays(year), fmt.Errorf("holiday API returned non-200 status code: %d", resp.StatusCode)
	}

	var apiHolidays []APIHoliday
	if err := json.NewDecoder(resp.Body).Decode(&apiHolidays); err != nil {
		return getFallbackHolidays(year), fmt.Errorf("failed to decode holiday API response: %w", err)
	}

	var holidays []Holiday
	for _, h := range apiHolidays {
		if !h.IsNationalHoliday {
			continue
		}
		holidayTime, err := time.Parse("2006-01-02", h.HolidayDate)
		if err != nil {
			// Skip invalid dates
			continue
		}
		if year != 0 && holidayTime.Year() != year {
			continue
		}
		holidays = append(holidays, Holiday{
			Date: holidayTime.Format("2006-01-02"),
			Name: h.HolidayName,
		})
	}

	return holidays, nil
}

// getFallbackHolidays returns a hardcoded list of holidays for a given year.
func getFallbackHolidays(year int) []Holiday {
	if year == 0 {
		year = time.Now().Year()
	}
	return []Holiday{
		{Date: fmt.Sprintf("%d-01-01", year), Name: "Tahun Baru Masehi"},
		{Date: fmt.Sprintf("%d-05-01", year), Name: "Hari Buruh Internasional"},
		{Date: fmt.Sprintf("%d-06-01", year), Name: "Hari Lahir Pancasila"},
		{Date: fmt.Sprintf("%d-08-17", year), Name: "Hari Kemerdekaan RI"},
		{Date: fmt.Sprintf("%d-12-25", year), Name: "Hari Raya Natal"},
	}
}
