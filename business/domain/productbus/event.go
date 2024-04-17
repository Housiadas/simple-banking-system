package productbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Housiadas/backend-system/business/domain/userbus"
	"github.com/Housiadas/backend-system/business/sys/delegate"
)

// registerDelegateFunctions will register action functions with the delegate
// system. If the core was constructed for query only, there won't be a
// delegate provided.
func (c *Core) registerDelegateFunctions() {
	if c.delegate != nil {
		c.delegate.Register(userbus.Domain, userbus.ActionUpdated, c.actionUserUpdated)
	}
}

// actionUserUpdated is executed by the user domain indirectly when a user is updated.
func (c *Core) actionUserUpdated(ctx context.Context, data delegate.Data) error {
	var params userbus.ActionUpdatedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	c.log.Info(ctx, "action user-update", "user_id", params.UserID, "enabled", params.Enabled)

	// Now we can see if this user has been disabled. If they have been, we will
	// want to disable to mark all these products as deleted. Right now we don't
	// have support for this, but you can see how we can process the event.

	return nil
}
