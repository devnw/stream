package stream

import "context"

// defaultCtx is the default context used by the stream package. This is
// hardcoded to context.Background() but can be overridden by the unit tests.
var defaultCtx = context.Background()

// _ctx returns a valid Context with CancelFunc even if it the
// supplied context is initially nil. If the supplied context
// is nil it uses the default context.
func _ctx(c context.Context) (context.Context, context.CancelFunc) {
	if c == nil {
		c = defaultCtx
	}

	return context.WithCancel(c)
}
