package wechat

import (
	"fmt"

	"github.com/medivhzhan/weapp/v2"
)

// Service implements a wechat auth service
type Service struct {
	AppID     string
	AppSecret string
}

// Resolve resolves authorization code to wechat open id
func (s *Service) Resolve(code string) (string, error) {
	res, err := weapp.Login(s.AppID, s.AppSecret, code)
	if err != nil {
		return "", fmt.Errorf("weapp.Login: %v", err)
	}

	if err := res.GetResponseError(); err != nil {
		return "", fmt.Errorf("weapp response error: %v", err)
	}

	return res.OpenID, nil
}
