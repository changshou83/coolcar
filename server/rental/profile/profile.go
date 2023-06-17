package profile

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityResolver interface {
	Resolve(c context.Context, photo []byte) (*rentalpb.Identity, error)
}

type Service struct {
	IdentityResolver  IdentityResolver
	BlobClient        blobpb.BlobServiceClient
	PhotoGetExpire    time.Duration
	PhotoUploadExpire time.Duration

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

// GetProfilePhoto returns a url for download files.
func (s *Service) GetProfilePhoto(
	c context.Context,
	req *rentalpb.GetProfilePhotoRequest,
) (*rentalpb.GetProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	profile, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err), "")
	}
	if profile.PhotoBlobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	record, err := s.BlobClient.GetBlobURL(c, &blobpb.GetBlobURLRequest{
		Id:         profile.PhotoBlobID,
		TimeoutSec: int32(s.PhotoGetExpire.Seconds()),
	})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &rentalpb.GetProfilePhotoResponse{
		Url: record.Url,
	}, nil
}

// CreateProfilePhoto returns a url for upload files.
func (s *Service) CreateProfilePhoto(
	c context.Context,
	req *rentalpb.CreateProfilePhotoRequest,
) (*rentalpb.CreateProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}
	// 获得预签名 URL
	record, err := s.BlobClient.CreateBlob(c, &blobpb.CreateBlobRequest{
		AccountId:           aid.String(),
		UploadUrlTimeoutSec: int32(s.PhotoUploadExpire.Seconds()),
	})
	if err != nil {
		s.Logger.Error("cannot create blob", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}
	// 为对应 profile 添加 blobid
	err = s.Mongo.UpdateProfilePhoto(c, aid, id.BlobID(record.Id))
	if err != nil {
		s.Logger.Error("cannot update profile photo", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}
	// 返回预签名 URL
	return &rentalpb.CreateProfilePhotoResponse{
		UploadUrl: record.UploadUrl,
	}, nil
}

// VerifyProfilePhoto returns AI recognition results.
func (s *Service) VerifyProfilePhoto(
	c context.Context,
	req *rentalpb.VerifyProfilePhotoRequest,
) (*rentalpb.Identity, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}
	// get blob id
	profile, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err), "")
	}
	if profile.PhotoBlobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}
	// get blob
	blob, err := s.BlobClient.GetBlob(c, &blobpb.GetBlobRequest{
		Id: profile.PhotoBlobID,
	})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}
	// verify
	s.Logger.Info("got profile photo", zap.Int("size", len(blob.Data)))
	return &rentalpb.Identity{
		LicNumber:   "210282198809294228",
		Name:        "王涛涛",
		Gender:      rentalpb.Gender_FEMALE,
		BirthDateMs: 631152000000,
	}, nil
	// return s.IdentityResolver.Resolve(c, blob.Data)
}

// ClearProfilePhoto clears profile photo.
func (s *Service) ClearProfilePhoto(
	c context.Context,
	req *rentalpb.ClearProfilePhotoRequest,
) (*rentalpb.ClearProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	err = s.Mongo.UpdateProfilePhoto(c, aid, id.BlobID(""))
	if err != nil {
		s.Logger.Error("cannot clear profile photo", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	return &rentalpb.ClearProfilePhotoResponse{}, nil
}

// logAndConvertProfileErr converts mongo err to http err
func (s *Service) logAndConvertProfileErr(err error) codes.Code {
	if err == mongo.ErrNoDocuments {
		return codes.NotFound
	}
	s.Logger.Error("cannot get profile", zap.Error(err))
	return codes.Internal
}
