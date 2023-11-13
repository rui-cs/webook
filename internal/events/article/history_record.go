package article

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/pkg/logger"
	"github.com/rui-cs/webook/pkg/saramax"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewHistoryReadEventConsumer(
	client sarama.Client,
	l logger.LoggerV1,
	repo repository.InteractiveRepository) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (r *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (r *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.AddRecord(ctx, t.Aid, t.Uid)
}
