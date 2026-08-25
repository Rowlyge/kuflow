package apikey

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {

	keys, err := h.service.List(
		r.Context(),
	)
	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(keys); err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}
}
