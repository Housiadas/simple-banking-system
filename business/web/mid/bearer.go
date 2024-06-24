package mid

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/Housiadas/backend-system/business/sys/errs"
	"github.com/Housiadas/backend-system/business/web"
)

// Bearer processes JWT token.
func (m *Mid) Bearer() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			claims, err := m.Bus.Auth.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				err = errs.New(errs.Unauthenticated, err)
				m.Log.Error(ctx, "bearer mid: unauthenticated", errs.Unauthenticated)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					return
				}
				return
			}

			if claims.Subject == "" {
				m.Log.Info(ctx, "request unauthenticated", errs.Unauthenticated)
				err = errs.New(errs.Unauthenticated, errors.New("authorize: you are not authorized for that action, no claims"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					return
				}
				return
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				m.Log.Error(ctx, "bearer mid: parsing", errs.Newf(errs.Unauthenticated, "parsing subject: %s", err))
				return
			}

			ctx = web.SetUserID(ctx, subjectID)
			ctx = web.SetClaims(ctx, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
