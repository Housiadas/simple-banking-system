package http

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Housiadas/simple-banking-system/foundation/metrics"
	"github.com/Housiadas/simple-banking-system/foundation/web"
)

// Panics recovers from panics and converts the panic to an error, so it is
// reported in Metrics and handled in Errors.
func Panics() web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

					metrics.AddPanics(ctx)
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
