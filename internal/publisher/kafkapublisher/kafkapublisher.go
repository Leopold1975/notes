package kafkapublisher

import (
	"notes/internal/pkg/config"
	"notes/internal/pkg/messaging/kafkabroker"
)

func New(cfg config.Kafka) (*kafkabroker.KafkaBroker, error) {
	kb := kafkabroker.New(cfg)
	err := kb.RegisterKafkaWriter()
	if err != nil {
		return &kafkabroker.KafkaBroker{}, err
	}

	return kb, nil
}
