package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Housiadas/backend-system/foundation/logger"
	"github.com/Housiadas/backend-system/foundation/tracer"
)

// HandlerFunc represents a function that handles a http request
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) (Encoder, error)

// Encoder defines behavior that can encode a data model and provide the content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

type httpStatus interface {
	HTTPStatus() int
}

type Respond struct {
	Log *logger.Logger
}

func NewRespond(log *logger.Logger) *Respond {
	return &Respond{
		Log: log,
	}
}

func (respond *Respond) Respond(handlerFunc HandlerFunc) http.HandlerFunc {
	// This is the decorator/middleware pattern in golang
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := SetTraceID(r.Context(), uuid.NewString())

		// Executes the handlerFunc for the specific route
		resp, err := handlerFunc(ctx, w, r)
		if err != nil {
			if err := responseError(ctx, w, err); err != nil {
				respond.Log.Error(ctx, "web-respond-error", "ERROR", err)
			}
			return
		}

		if err := responseData(ctx, w, resp); err != nil {
			respond.Log.Error(ctx, "web-respond", "ERROR", err)
		}
	}
}

func responseError(ctx context.Context, w http.ResponseWriter, err error) error {
	data, ok := err.(Encoder)
	if !ok {
		return fmt.Errorf("error value does not implement the encoder interface: %T", err)
	}

	return responseData(ctx, w, data)
}

func responseData(ctx context.Context, w http.ResponseWriter, dataModel Encoder) error {
	// If the context has been canceled, it means the client is no longer waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	var statusCode = http.StatusOK

	switch v := dataModel.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()

	case error:
		statusCode = http.StatusInternalServerError

	default:
		if dataModel == nil {
			statusCode = http.StatusNoContent
		}
	}

	_, span := tracer.AddSpan(ctx, "web.response", attribute.Int("status", statusCode))
	defer span.End()

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := dataModel.Encode()
	if err != nil {
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
