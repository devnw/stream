package stream

import "context"

// defaultCtx is the default context used by the stream package. This is
// hardcoded to context.Background() but can be overridden by the unit tests.
//
//nolint:gochecknoglobals // this is on purpose
var defaultCtx = context.Background()

// _ctx returns a valid Context with CancelFunc even if it the
// supplied context is initially nil. If the supplied context
// is nil it uses the default context.
func _ctx(c context.Context) context.Context {
	if c == nil {
		c = defaultCtx
	}

	return c
}
