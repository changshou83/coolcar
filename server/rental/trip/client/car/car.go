package car

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
)

type Manager struct {
}

func (m *Manager) Verify(context.Context, id.CarID, *rentalpb.Location) error {
	return nil
}

func (m *Manager) Unlock(c context.Context, cid id.CarID, aid id.AccountID, tid id.TripID, avatarURL string) error {
	return nil
}

func (m *Manager) Lock(c context.Context, cid id.CarID) error {
	return nil
}
