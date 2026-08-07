package proxy

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/upstream"
)

type upstreamKey struct{}
type upstreamUnavailableKey struct{}

// IntoContext сохраняет upstream в Context.
func IntoContext(
	ctx context.Context,
	up *upstream.Upstream,
) context.Context {

	return context.WithValue(
		ctx,
		upstreamKey{},
		up,
	)
}

// UpstreamFromContext возвращает upstream.
func UpstreamFromContext(
	ctx context.Context,
) *upstream.Upstream {

	up, ok := ctx.Value(
		upstreamKey{},
	).(*upstream.Upstream)

	if !ok {
		return nil
	}

	return up
}

// UpstreamNameFromContext возвращает имя upstream.
func UpstreamNameFromContext(
	ctx context.Context,
) string {

	up := UpstreamFromContext(ctx)

	if up == nil {
		return ""
	}

	return up.Name
}

// MarkUpstreamUnavailable помечает запрос,
// для которого не удалось выбрать доступный upstream.
func MarkUpstreamUnavailable(
	ctx context.Context,
) context.Context {

	return context.WithValue(
		ctx,
		upstreamUnavailableKey{},
		true,
	)
}

// IsUpstreamUnavailable проверяет,
// не удалось ли выбрать доступный upstream.
func IsUpstreamUnavailable(
	ctx context.Context,
) bool {

	value, ok := ctx.Value(
		upstreamUnavailableKey{},
	).(bool)

	return ok && value
}
