package profile

import (
	"context"
	"coolcar/shared/id"
)

type Manager struct {
}

func (m *Manager) Verify(
	c context.Context,
	aid id.AccountID,
) (id.IdentityID, error) {
	return id.IdentityID("identity1"), nil
}
