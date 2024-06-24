package authbus

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Authenticate processes the token to validation the sender's token is valid.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	var claims Claims
	token, _, err := a.parser.ParseUnverified(parts[1], &claims)
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	kidRaw, exists := token.Header["kid"]
	if !exists {
		return Claims{}, fmt.Errorf("kid missing from header: %w", err)
	}

	kid, ok := kidRaw.(string)
	if !ok {
		return Claims{}, fmt.Errorf("kid malformed: %w", err)
	}

	pem, err := a.keyLookup.PublicKey(kid)
	if err != nil {
		return Claims{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": parts[1],
		"ISS":   a.issuer,
	}

	if err := a.opaPolicyEvaluation(ctx, regoAuthentication, RuleAuthenticate, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed : %w", err)
	}

	// Check the database for this user to verify they are still enabled.

	if err := a.isUserEnabled(ctx, claims); err != nil {
		return Claims{}, fmt.Errorf("user not enabled : %w", err)
	}

	return claims, nil
}
