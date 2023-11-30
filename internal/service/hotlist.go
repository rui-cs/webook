package service

import (
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository"
)

type HotListService interface {
	GetLikeTopN(bizs []string) (map[string][]domain.HotList, error)
}

type hotListService struct {
	repo repository.HotListRepo
}

func NewHotListService(repo repository.HotListRepo) HotListService {
	return &hotListService{repo: repo}
}

func (h *hotListService) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	return h.repo.GetLikeTopN(bizs)
}
