package auth

import "errors"

var (
	ErrMissingAPIKey  = errors.New("missing api key")
	ErrInvalidAPIKey  = errors.New("invalid api key")
	ErrDisabledAPIKey = errors.New("api key is disabled")
	ErrExpiredAPIKey  = errors.New("api key is expired")
)
