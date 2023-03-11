package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// type IdentityResolver interface {
// 	Resolver(c context.Context, photo []byte) (*rentalpb.Identity, error)
// }

type Service struct {
	// IdentityResolver IdentityResolver
	Mongo  *dao.Mongo
	Logger *zap.Logger
	rentalpb.UnimplementedProfileServiceServer
}

func (s *Service) GetProfile(
	c context.Context,
	req *rentalpb.GetProfileRequest,
) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	record, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		code := s.logAndConvertProfileErr(err)
		if code == codes.NotFound {
			return &rentalpb.Profile{}, nil
		}
		return nil, status.Error(code, "")
	}
	if record.Profile == nil {
		return &rentalpb.Profile{}, nil
	}

	return record.Profile, nil
}

func (s *Service) SubmitProfile(
	c context.Context,
	identity *rentalpb.Identity,
) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	profile := &rentalpb.Profile{
		Identity: identity,
		Status:   rentalpb.IdentityStatus_PENDING,
	}
	err = s.Mongo.UpdateProfile(c, aid, rentalpb.IdentityStatus_UNSUBMITTED, profile)
	if err != nil {
		s.Logger.Error("cannot submit profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	// 默认三秒钟之后通过验证
	go func() {
		time.Sleep(3 * time.Second)
		err := s.Mongo.UpdateProfile(
			context.Background(),
			aid,
			rentalpb.IdentityStatus_PENDING,
			&rentalpb.Profile{
				Identity: identity,
				Status:   rentalpb.IdentityStatus_VERIFIED,
			},
		)
		if err != nil {
			s.Logger.Error("cannot verify identity", zap.Error(err))
		}
	}()
	return profile, nil
}

// ClearProfile clears verified profile for an account.
func (s *Service) ClearProfile(
	c context.Context,
	req *rentalpb.ClearProfileRequest,
) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	profile := &rentalpb.Profile{}
	err = s.Mongo.UpdateProfile(c, aid, rentalpb.IdentityStatus_VERIFIED, profile)
	if err != nil {
		s.Logger.Error("cannot clear profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	return profile, nil
}

// logAndConvertProfileErr converts mongo err to http err
func (s *Service) logAndConvertProfileErr(err error) codes.Code {
	if err == mongo.ErrNoDocuments {
		return codes.NotFound
	}
	s.Logger.Error("cannot get profile", zap.Error(err))
	return codes.Internal
}
