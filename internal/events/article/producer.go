package article

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
	ProduceReadEventV1(ctx context.Context, v1 ReadEventV1)
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func (k *KafkaProducer) ProduceReadEventV1(ctx context.Context, v1 ReadEventV1) {
	//TODO implement me
	panic("implement me")
}

// ProduceReadEvent 如果你有复杂的重试逻辑，就用装饰器
// 你认为你的重试逻辑很简单，你就放这里
func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func NewKafkaProducer(pc sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: pc,
	}
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type ReadEventV1 struct {
	Uids []int64
	Aids []int64
}
