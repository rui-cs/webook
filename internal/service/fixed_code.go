package service

import (
	"context"

	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/service/sms"
)

// 固定验证码，接口测试用
type fixedCodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewFixedCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &fixedCodeService{repo: repo, smsSvc: smsSvc}
}

func (s *fixedCodeService) Send(ctx context.Context, biz, phone string) error {
	code := s.generateCode()
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	return s.smsSvc.Send(ctx, "", []string{code}, phone)
}

func (s *fixedCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return s.repo.Verify(ctx, biz, phone, inputCode)
}

func (s *fixedCodeService) generateCode() string {
	return "207391"
}
