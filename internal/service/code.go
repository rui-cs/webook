package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/service/sms"
	"go.uber.org/atomic"
)

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeOperationTooMany   = repository.ErrCodeOperationTooMany
)

var codeTplId atomic.String = atomic.String{}

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	codeTplId.Store("1877556")
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	codeTplId.Store(viper.GetString("code.tpl.id"))
	//})

	return &codeService{repo: repo, smsSvc: smsSvc}
}

func (s *codeService) Send(ctx context.Context, biz, phone string) error {
	code := s.generateCode()
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	err = s.smsSvc.Send(ctx, codeTplId.Load(), []string{code}, phone)

	if err != nil {
		err = fmt.Errorf("发送短信出现异常 %w", err)
	}

	return err
}

func (s *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return s.repo.Verify(ctx, biz, phone, inputCode)
}

func (s *codeService) generateCode() string {
	num := rand.Intn(1000000) //[0-999999]
	return fmt.Sprintf("%06d", num)
}
