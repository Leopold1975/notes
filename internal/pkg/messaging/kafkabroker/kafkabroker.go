package kafkabroker

import (
	"context"
	"fmt"
	"net"
	"notes/internal/pkg/config"
	"notes/internal/pkg/models"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaBroker struct {
	cfg    config.Kafka
	conn   *kafka.Conn
	writer *kafka.Writer
	reader *kafka.Reader
}

func New(cfg config.Kafka) *KafkaBroker {
	return &KafkaBroker{
		cfg: cfg,
	}
}

func (kb *KafkaBroker) RegisterKafkaWriter() error {
	if err := kb.connect(kb.cfg); err != nil {
		return err
	}
	kb.writer = &kafka.Writer{
		Addr:     kafka.TCP(kb.cfg.Brokers...),
		Topic:    kb.cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	return nil
}

func (kb *KafkaBroker) RegisterKafkaReader() error {
	if err := kb.connect(kb.cfg); err != nil {
		return err
	}
	kb.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     kb.cfg.Brokers,
		Topic:       kb.cfg.Topic,
		MaxBytes:    10e6,
		GroupID:     kb.cfg.Group,
		StartOffset: kafka.FirstOffset,
	})
	return nil
}

func (kb *KafkaBroker) Send(ctx context.Context, m models.Message) error {
	if kb.writer == nil {
		err := kb.RegisterKafkaWriter()
		if err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	return kb.writer.WriteMessages(ctx, kafka.Message{
		Key:   m.Key,
		Value: m.Value,
	})
}

func (kb *KafkaBroker) Receive(ctx context.Context) (models.Message, error) {
	if kb.reader == nil {
		if err := kb.RegisterKafkaReader(); err != nil {
			return models.Message{}, err
		}
	}
	m, err := kb.reader.ReadMessage(ctx)
	if err != nil {
		return models.Message{}, err
	}

	var mm models.Message
	mm.Key = m.Key
	mm.Value = m.Value

	return mm, nil
}

func (kb *KafkaBroker) Shutdown() error {
	var errW, errR error
	if kb.writer != nil {
		errW = kb.writer.Close()
	}
	if kb.reader != nil {
		errR = kb.reader.Close()
	}

	errC := kb.conn.Close()

	var errMessage string
	if errW != nil {
		errMessage += "error closing writer: %w\n"
	}
	if errR != nil {
		errMessage += "error closing reader: %w\n"
	}
	if errC != nil {
		errMessage += "error closing connection: %w\n"
	}
	return fmt.Errorf(errMessage, errW, errR, errC)
}

func (kb *KafkaBroker) connect(cfg config.Kafka) error {
	n := 1e9

	conn, err := kafka.Dial("tcp", cfg.Brokers[0])
	for err != nil {
		time.Sleep(time.Duration(n))
		n += 3e9
		conn, err = kafka.Dial("tcp", cfg.Brokers[0])
		if n < 20e9 {
			continue
		}
		return err
	}
	defer conn.Close()

	contr, err := conn.Controller()
	if err != nil {
		return err
	}

	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(contr.Host, strconv.Itoa(contr.Port)))
	for err != nil {
		time.Sleep(time.Duration(n))
		n += 3e9
		controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(contr.Host, strconv.Itoa(contr.Port)))
		if n < 20e9 {
			continue
		}
		return err
	}
	kb.conn = controllerConn

	topicCfg := []kafka.TopicConfig{
		{
			Topic:             cfg.Topic,
			NumPartitions:     cfg.Partitions,
			ReplicationFactor: kb.cfg.ReplicationFactor,
		},
	}

	err = conn.CreateTopics(topicCfg...)
	if err != nil {
		return err
	}
	return nil
}
