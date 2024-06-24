package mid

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/Housiadas/backend-system/business/domain/authbus"
	"github.com/Housiadas/backend-system/business/domain/productbus"
	"github.com/Housiadas/backend-system/business/domain/userbus"
	"github.com/Housiadas/backend-system/business/sys/errs"
	"github.com/Housiadas/backend-system/business/web"
)

// ErrInvalidID represents a condition where the id is not an uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize validates authorization.
func (m *Mid) Authorize(rule string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, err := web.GetUserID(ctx)
			if err != nil {
				err = errs.New(errs.Unauthenticated, err)
				m.Log.Error(ctx, "authorize mid: get user id", err)
				http.Error(w, err.Error(), errs.Unauthenticated.Value())
				return
			}

			authData := authbus.Authorize{
				Claims: web.GetClaims(ctx),
				UserID: userID,
				Rule:   rule,
			}

			if err := m.Bus.Auth.Authorize(ctx, authData.Claims, authData.UserID, authData.Rule); err != nil {
				err = errs.New(errs.Unauthenticated, err)
				m.Log.Error(ctx, "authorize mid: authorize", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					return
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorizeUser executes the specified role and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified user id.
func (m *Mid) AuthorizeUser(rule string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID uuid.UUID
			id := web.Param(r, "user_id")
			ctx := r.Context()

			if id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					err = errs.New(errs.Unauthenticated, ErrInvalidID)
					m.Log.Error(ctx, "authorize user mid: authorize", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					if err := json.NewEncoder(w).Encode(err); err != nil {
						return
					}
					return
				}

				usr, err := m.Bus.User.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, userbus.ErrNotFound):
						err = errs.New(errs.Unauthenticated, err)
					default:
						err = errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
					}
					m.Log.Error(ctx, "authorize user mid: authorize", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					if err := json.NewEncoder(w).Encode(err); err != nil {
						return
					}
					return
				}

				ctx = web.SetUser(ctx, usr)
			}

			authData := authbus.Authorize{
				Claims: web.GetClaims(ctx),
				UserID: userID,
				Rule:   rule,
			}

			if err := m.Bus.Auth.Authorize(ctx, authData.Claims, authData.UserID, authData.Rule); err != nil {
				err = errs.New(errs.Unauthenticated, err)
				m.Log.Error(ctx, "authorize user mid: authorize", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					return
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorizeProduct executes the specified role and extracts the specified
// productapi from the DB if a productapi id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the productapi.
func (m *Mid) AuthorizeProduct(rule string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID uuid.UUID
			id := web.Param(r, "product_id")
			ctx := r.Context()

			if id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					err = errs.New(errs.Unauthenticated, ErrInvalidID)
					m.Log.Error(ctx, "authorize product mid: authorize", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					if err := json.NewEncoder(w).Encode(err); err != nil {
						return
					}
					return
				}

				prd, err := m.Bus.Product.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, productbus.ErrNotFound):
						err = errs.New(errs.Unauthenticated, err)
					default:
						err = errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
					}
					m.Log.Error(ctx, "authorize product mid: authorize", err)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					if err := json.NewEncoder(w).Encode(err); err != nil {
						return
					}
					return
				}

				userID = prd.UserID
				ctx = web.SetProduct(ctx, prd)
			}

			authData := authbus.Authorize{
				Claims: web.GetClaims(ctx),
				UserID: userID,
				Rule:   rule,
			}

			if err := m.Bus.Auth.Authorize(ctx, authData.Claims, authData.UserID, authData.Rule); err != nil {
				err = errs.New(errs.Unauthenticated, err)
				m.Log.Error(ctx, "authorize product mid: authorize", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					return
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
