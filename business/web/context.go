package web

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Housiadas/backend-system/business/domain/authbus"
	"github.com/Housiadas/backend-system/business/domain/productbus"
	"github.com/Housiadas/backend-system/business/domain/userbus"
)

type ctxKey int

const (
	requestID ctxKey = iota + 1
	apiVersion
	traceKey
	claimKey
	userIDKey
	userKey
	productKey
)

func SetRequestID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, requestID, reqId)
}

func GetRequestID(ctx context.Context) string {
	v, ok := ctx.Value(requestID).(string)
	if !ok {
		return ""
	}
	return v
}

func SetApiVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersion, version)
}

func GetApiVersion(ctx context.Context) string {
	v, ok := ctx.Value(apiVersion).(string)
	if !ok {
		return ""
	}
	return v
}

func SetClaims(ctx context.Context, claims authbus.Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) authbus.Claims {
	v, ok := ctx.Value(claimKey).(authbus.Claims)
	if !ok {
		return authbus.Claims{}
	}
	return v
}

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID returns the user id from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

func SetUser(ctx context.Context, usr userbus.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (userbus.User, error) {
	v, ok := ctx.Value(userKey).(userbus.User)
	if !ok {
		return userbus.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func SetProduct(ctx context.Context, prd productbus.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// GetProduct returns the productapi from the context.
func GetProduct(ctx context.Context) (productbus.Product, error) {
	v, ok := ctx.Value(productKey).(productbus.Product)
	if !ok {
		return productbus.Product{}, errors.New("productapi not found in context")
	}

	return v, nil
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey, traceID)
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(traceKey).(string)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}

	return v
}
