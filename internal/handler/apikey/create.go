package apikey

import (
	"encoding/json"
	"net/http"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
	"github.com/google/uuid"
)

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {

		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)

		return
	}

	key := &apikeyrepo.APIKey{
		APIKey:  uuid.NewString(),
		Owner:   request.Owner,
		Enabled: true,
	}

	if err := h.service.Create(
		r.Context(),
		key,
	); err != nil {

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

	w.WriteHeader(
		http.StatusCreated,
	)

	_ = json.NewEncoder(w).Encode(key)
}
