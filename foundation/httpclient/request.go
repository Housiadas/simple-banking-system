package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"go.opentelemetry.io/otel/attribute"

	"github.com/Housiadas/backend-system/business/sys/errs"
	"github.com/Housiadas/backend-system/foundation/tracer"
)

func (cln *Client) rawRequest(ctx context.Context, method string, endpoint string, headers map[string]string, r io.Reader, v any) error {
	var statusCode int

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint: %w", err)
	}
	base := path.Base(u.Path)

	ctx, span := tracer.AddSpan(ctx, fmt.Sprintf("app.api.authclient.%s", base), attribute.String("endpoint", endpoint))
	defer func() {
		span.SetAttributes(attribute.Int("status", statusCode))
		span.End()
	}()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := cln.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: error: %w", err)
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode

	if statusCode == http.StatusNoContent {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	switch statusCode {
	case http.StatusNoContent:
		return nil

	case http.StatusOK:
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return nil

	case http.StatusUnauthorized:
		var err *errs.Error
		if err := json.Unmarshal(data, &err); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return err

	default:
		return fmt.Errorf("failed: response: %s", string(data))
	}
}
