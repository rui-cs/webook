package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/service/sms"
)

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeOperationTooMany   = repository.ErrCodeOperationTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{repo: repo, smsSvc: smsSvc}
}

func (s *codeService) Send(ctx context.Context, biz, phone string) error {
	code := s.generateCode()
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	return s.smsSvc.Send(ctx, "", []string{code}, phone)
}

func (s *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return s.repo.Verify(ctx, biz, phone, inputCode)
}

func (s *codeService) generateCode() string {
	num := rand.Intn(1000000) //[0-999999]
	return fmt.Sprintf("%06d", num)
}
