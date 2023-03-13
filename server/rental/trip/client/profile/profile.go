package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Fetcher defines the interface to fetch profile.
type Fetcher interface {
	GetProfile(c context.Context, req *rentalpb.GetProfileRequest) (*rentalpb.Profile, error)
}

// Manager defines a profie manager.
type Manager struct {
	Fetcher Fetcher
}

// Verify verifies account identity.
func (m *Manager) Verify(
	c context.Context,
	aid id.AccountID,
) (id.IdentityID, error) {
	emptyID := id.IdentityID("")
	profile, err := m.Fetcher.GetProfile(c, &rentalpb.GetProfileRequest{})
	if err != nil {
		return emptyID, fmt.Errorf("cannot get profile: %v", err)
	}
	if profile.Status != rentalpb.IdentityStatus_VERIFIED {
		return emptyID, fmt.Errorf("invalid indentity status")
	}

	identity, err := proto.Marshal(profile.Identity)
	if err != nil {
		return emptyID, fmt.Errorf("cannot marshal identity: %v", err)
	}
	// 对驾驶者身份进行二进制编码作为 identity id，用于在唯一trip中标识驾驶者
	return id.IdentityID(base64.StdEncoding.EncodeToString(identity)), nil
}
