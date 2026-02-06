package handlers

import (
	"net/http"
	"strconv"
	"dsi_interna_sys/internal/holiday"
	"dsi_interna_sys/internal/utils"
)

func GetHolidays(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	var year int
	var err error
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			utils.RespondBadRequest(w, "Invalid year format")
			return
		}
	}

	holidays, err := holiday.GetHolidays(year)
	if err != nil {
		// If there is an error, we still return the fallback holidays
		utils.RespondSuccess(w, "Successfully retrieved fallback holidays", holidays)
		return
	}

	utils.RespondSuccess(w, "Successfully retrieved holidays", holidays)
}
