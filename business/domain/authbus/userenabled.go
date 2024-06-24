package authbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// isUserEnabled hits the database and checks the user is not disabled. If the
// no database connection was provided, this check is skipped.
func (a *Auth) isUserEnabled(ctx context.Context, claims Claims) error {
	if a.userBus == nil {
		return nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("parse user: %w", err)
	}

	usr, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}

	if !usr.Enabled {
		return fmt.Errorf("user disabled")
	}

	return nil
}
