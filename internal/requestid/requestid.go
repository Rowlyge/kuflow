package requestid

import (
	"context"

	"github.com/google/uuid"
)

// contextKey используется,
// чтобы избежать конфликтов ключей Context.
type contextKey string

const (
	headerName = "X-Request-ID"

	requestIDKey contextKey = "request-id"
)

// New создаёт новый Request ID.
func New() string {
	return uuid.NewString()
}

// FromHeader возвращает Request ID,
// переданный клиентом, либо создаёт новый.
func FromHeader(headerValue string) string {

	if headerValue != "" {
		return headerValue
	}

	return New()
}

// IntoContext сохраняет Request ID в Context.
func IntoContext(
	ctx context.Context,
	id string,
) context.Context {

	return context.WithValue(
		ctx,
		requestIDKey,
		id,
	)
}

// FromContext получает Request ID из Context.
func FromContext(
	ctx context.Context,
) string {

	id, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return id
}

// HeaderName возвращает имя HTTP-заголовка.
func HeaderName() string {
	return headerName
}
