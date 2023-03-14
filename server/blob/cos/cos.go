package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type Service struct {
	client    *cos.Client
	secretID  string
	secretKey string
}

// NewService creates a cos services.
func NewService(addr, secretID, secretKey string) (*Service, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse addr: %v", err)
	}

	baseURL := &cos.BaseURL{BucketURL: url}

	return &Service{
		client: cos.NewClient(
			baseURL,
			&http.Client{
				Transport: &cos.AuthorizationTransport{
					SecretID:  secretID,
					SecretKey: secretKey,
				},
			},
		),
		secretID:  secretID,
		secretKey: secretKey,
	}, nil
}

// SignURL signs a url for download or upload files.
func (s *Service) SignURL(c context.Context, method, path string, timeout time.Duration) (string, error) {
	url, err := s.client.Object.GetPresignedURL(c, method, path, s.secretID, s.secretKey, timeout, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// Get gets storage contents.
func (s *Service) Get(c context.Context, path string) (io.ReadCloser, error) {
	res, err := s.client.Object.Get(c, path, nil)

	var b io.ReadCloser
	if res != nil {
		b = res.Body
	}
	if err != nil {
		return b, err
	}
	if res.StatusCode >= 400 {
		return b, fmt.Errorf("got err response: %+v", res)
	}

	return b, nil
}
