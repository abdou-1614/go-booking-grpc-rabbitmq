package kafka

import (
	"Go-grpc/config"
	"Go-grpc/internal/model"
	"Go-grpc/internal/user"
	"Go-grpc/pkg/logger"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
)

type UserConsumerGroup struct {
	Brokers  []string
	GroupID  string
	cfg      *config.Config
	logger   logger.Loggor
	userUC   user.UseCase
	validate *validator.Validate
}

func NewUserConsumerGroup(Brokers []string, GroupID string, cfg *config.Config, logger logger.Loggor, user user.UseCase, validate *validator.Validate) *UserConsumerGroup {
	return &UserConsumerGroup{
		Brokers:  Brokers,
		GroupID:  GroupID,
		cfg:      cfg,
		logger:   logger,
		userUC:   user,
		validate: validate,
	}
}

func (u *UserConsumerGroup) getNewKafkaConsumer(brokers []string, groupID string, topic string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	// config.Consumer.Offsets.AutoCommit.Retry = 3

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)

	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func (u *UserConsumerGroup) getNewKafkaSyncProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(u.cfg.Kafka.Brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func (u *UserConsumerGroup) CreateUserConsumer(
	ctx context.Context,
	cancel context.CancelFunc,
	groupID string,
	topic string,
	workersNum int,
) {
	consumer, err := u.getNewKafkaConsumer(u.cfg.Kafka.Brokers, groupID, topic)

	if err != nil {
		u.logger.Errorf("Error creating Kafka consumer: %v", err)
		return
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			u.logger.Errorf("Error closing Kafka consumer: %v", err)
		}
	}()

	producer, err := u.getNewKafkaSyncProducer()

	if err != nil {
		u.logger.Errorf("Error creating Kafka producer: %v", err)
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			u.logger.Errorf("Error closing Kafka producer: %v", err)
		}
	}()

	u.logger.Infof("Starting consumer group: %v", groupID)

	wg := &sync.WaitGroup{}

	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go u.createUserWorker(ctx, cancel, consumer, producer, wg, i)
	}
	wg.Wait()
}

func (u *UserConsumerGroup) publishErrorMessage(ctx context.Context, w sarama.SyncProducer, m *sarama.ConsumerMessage, err error) error {
	errMsg := model.ErrorMessage{
		Offset:    m.Offset,
		Error:     err.Error(),
		Time:      m.Timestamp.UTC(),
		Partition: m.Partition,
		Topic:     m.Topic,
	}
	errMsgBytes, err := json.Marshal(errMsg)
	if err != nil {
		return err
	}

	_, _, err = w.SendMessage(&sarama.ProducerMessage{
		Topic: deadLetterQueueTopic,
		Value: sarama.StringEncoder(errMsgBytes),
	})

	if err != nil {
		return err
	}

	return nil
}

func (u *UserConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go u.CreateUserConsumer(ctx, cancel, u.GroupID, CreateUserTopic, CreateUsertWorkers)
}
