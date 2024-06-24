package authbus

import (
	"encoding/json"

	"github.com/google/uuid"
)

type AuthenticateResp struct {
	Token string `json:"token"`
}

// Encode implements the encoder interface.
func (a AuthenticateResp) Encode() ([]byte, string, error) {
	data, err := json.Marshal(a)
	return data, "application/json", err
}

// Authorize defines the information required to perform an authorization.
type Authorize struct {
	Claims Claims
	UserID uuid.UUID
	Rule   string
}

// Error represents an error in the systemapi.
type Error struct {
	Message string `json:"message"`
}

// Error implements the error interface.
func (err Error) Error() string {
	return err.Message
}
