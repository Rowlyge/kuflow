package apikey

import (
	apikeyservice "github.com/Rowlyge/kuflow/internal/service/apikey"
)

type Handler struct {
	service *apikeyservice.Service
}

func New(
	service *apikeyservice.Service,
) *Handler {

	return &Handler{
		service: service,
	}
}
