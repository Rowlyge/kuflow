package apikey

import (
	"encoding/json"
	"net/http"

	apikeyservice "github.com/Rowlyge/kuflow/internal/service/apikey"
)

type StatsHandler struct {
	stats *apikeyservice.StatsService
}

func NewStatsHandler(
	stats *apikeyservice.StatsService,
) *StatsHandler {

	return &StatsHandler{
		stats: stats,
	}
}

func (h *StatsHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {

	stats, err := h.stats.List(
		r.Context(),
	)
	if err != nil {

		http.Error(
			w,
			"failed to load api key stats",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(stats); err != nil {

		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)

		return
	}
}
