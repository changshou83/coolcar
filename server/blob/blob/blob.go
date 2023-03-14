package blob

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/dao"
	"coolcar/shared/id"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Storage interface {
	SignURL(c context.Context, method, path string, timeout time.Duration) (string, error)
	Get(c context.Context, path string) (io.ReadCloser, error)
}

// Service defines a blob service.
type Service struct {
	Storage Storage
	Mongo   *dao.Mongo
	Logger  *zap.Logger
	blobpb.UnimplementedBlobServiceServer
}

func (s *Service) CreateBlob(c context.Context, req *blobpb.CreateBlobRequest) (*blobpb.CreateBlobResponse, error) {
	aid := id.AccountID(req.AccountId)
	record, err := s.Mongo.CreateBlob(c, aid)
	if err != nil {
		return nil, status.Error(codes.Internal, "")
	}

	url, err := s.Storage.SignURL(c, http.MethodPut, record.Path, sec2Duration(req.UploadUrlTimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign url: %v", err)
	}

	return &blobpb.CreateBlobResponse{
		Id:        record.ID.Hex(),
		UploadUrl: url,
	}, nil
}

// GetBlob gets a blob's contents.
func (s *Service) GetBlob(c context.Context, req *blobpb.GetBlobRequest) (*blobpb.GetBlobResponse, error) {
	record, err := s.getBlobRecord(c, id.BlobID(req.Id))
	if err != nil {
		return nil, err
	}

	reader, err := s.Storage.Get(c, record.Path)
	if reader != nil {
		defer reader.Close()
	}
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot get storage: %v", err)
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot read from response: %v", err)
	}

	return &blobpb.GetBlobResponse{
		Data: bytes,
	}, nil
}

// GetBlobURL gets blob's url for downloading.
func (s *Service) GetBlobURL(c context.Context, req *blobpb.GetBlobURLRequest) (*blobpb.GetBlobURLResponse, error) {
	record, err := s.getBlobRecord(c, id.BlobID(req.Id))
	if err != nil {
		return nil, err
	}

	url, err := s.Storage.SignURL(c, http.MethodGet, record.Path, sec2Duration(req.TimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign url: %v", err)
	}

	return &blobpb.GetBlobURLResponse{
		Url: url,
	}, nil
}

// getBlobRecord gets blob record from mongodb.
func (s *Service) getBlobRecord(c context.Context, bid id.BlobID) (*dao.BlobRecord, error) {
	record, err := s.Mongo.GetBlob(c, bid)
	if err == mongo.ErrNoDocuments {
		return nil, status.Error(codes.NotFound, "")
	}

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return record, nil
}

func sec2Duration(sec int32) time.Duration {
	return time.Duration(sec) * time.Second
}
