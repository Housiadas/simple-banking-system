package authbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Authorize attempts to authorize the user with the provided input roles, if
// none of the input roles are within the user's claims, we return an error
// otherwise the user is authorized.
func (a *Auth) Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error {
	input := map[string]any{
		"Roles":   claims.Roles,
		"Subject": claims.Subject,
		"UserID":  userID,
	}

	if err := a.opaPolicyEvaluation(ctx, regoAuthorization, rule, input); err != nil {
		return fmt.Errorf("rego evaluation failed : %w", err)
	}

	return nil
}
