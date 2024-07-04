package kafka

import (
	"Go-grpc/config"
	"Go-grpc/pkg/logger"
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successPublisherhMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rabbitmq_images_success_publish_messages_total",
		Help: "The total number of success RabbitMQ published messages",
	})

	errorPublisherMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rabbitmq_images_error_publish_messages_total",
		Help: "The total number of errors RabbitMq published messages",
	})
)

type UserProducer interface {
	PublishCreate(ctx context.Context, msgs ...*sarama.ProducerMessage) error
	PublishUpdate(ctx context.Context, msgs ...*sarama.ProducerMessage) error
	Close()
	Run()
	GetNewKafkaAsyncProducer(topic string) sarama.AsyncProducer
}

type userProducer struct {
	log               logger.Loggor
	cfg               *config.Config
	createAsyncWriter sarama.AsyncProducer
	updateAsyncWriter sarama.AsyncProducer
}

func NewUserProducer(log logger.Loggor, cfg *config.Config) *userProducer {
	return &userProducer{
		log: log,
		cfg: cfg,
	}
}

func (u *userProducer) GetNewKafkaAsyncProducer(topic string) sarama.AsyncProducer {
	asyncProducer, err := sarama.NewAsyncProducer(u.cfg.Kafka.Brokers, u.producerConfing())
	if err != nil {
		u.log.Fatalf("Error creating Kafka async producer: %s", err)
	}
	return asyncProducer
}

func (u *userProducer) Run() {
	u.createAsyncWriter = u.GetNewKafkaAsyncProducer(CreateUserTopic)
	u.updateAsyncWriter = u.GetNewKafkaAsyncProducer(UpdateUserTopic)
}

func (u *userProducer) Close() {
	_ = u.createAsyncWriter.Close()
	_ = u.updateAsyncWriter.Close()
}

func (p *userProducer) PublishCreate(ctx context.Context, msgs ...*sarama.ProducerMessage) error {
	for _, msg := range msgs {
		p.createAsyncWriter.Input() <- msg
	}
	return nil
}

// PublishUpdate publish messages to update topic
func (p *userProducer) PublishUpdate(ctx context.Context, msgs ...*sarama.ProducerMessage) error {
	for _, msg := range msgs {
		p.updateAsyncWriter.Input() <- msg
	}
	return nil
}

func (u *userProducer) producerConfing() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 40 * time.Second
	return config
}
