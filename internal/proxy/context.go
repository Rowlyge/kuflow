package proxy

import "context"

type contextKey string

const upstreamKey contextKey = "upstream"

func WithUpstream(
	ctx context.Context,
	name string,
) context.Context {

	return context.WithValue(
		ctx,
		upstreamKey,
		name,
	)
}

func UpstreamFromContext(
	ctx context.Context,
) string {

	value := ctx.Value(upstreamKey)
	if value == nil {
		return "unknown"
	}

	name, ok := value.(string)
	if !ok {
		return "unknown"
	}

	return name
}
