package apikey

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) Disable(
	w http.ResponseWriter,
	r *http.Request,
) {

	idPart := strings.TrimPrefix(
		r.URL.Path,
		"/admin/api-keys/",
	)

	idPart = strings.TrimSuffix(
		idPart,
		"/disable",
	)

	id, err := strconv.ParseInt(
		idPart,
		10,
		64,
	)
	if err != nil {

		http.Error(
			w,
			"invalid id",
			http.StatusBadRequest,
		)

		return
	}

	if err := h.service.Disable(
		r.Context(),
		id,
	); err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(
		http.StatusNoContent,
	)
}
