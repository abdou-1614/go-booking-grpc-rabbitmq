package kafka

import (
	"Go-grpc/config"
	"Go-grpc/internal/model"
	"Go-grpc/pkg/logger"
	"context"
	"encoding/json"
	"sync"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
)

func (ucg *UserConsumerGroup) createUserWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	consumer sarama.ConsumerGroup,
	producer sarama.SyncProducer,
	wg *sync.WaitGroup,
	workerID int,
) {
	defer wg.Done()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		default:

			handler := newUserHandler(ucg.logger, producer, ucg.cfg, *ucg.validate, workerID)
			if err := consumer.Consume(ctx, []string{CreateUserTopic}, handler); err != nil {
				ucg.logger.Errorf("Error consuming messages: %v", err)
			}
		}
	}
}

func newUserHandler(logger logger.Loggor, producer sarama.SyncProducer, cfg *config.Config, validate validator.Validate, workerID int) *userHandler {
	return &userHandler{
		log:      logger,
		producer: producer,
		cfg:      cfg,
		validate: validate,
		workerID: workerID,
	}
}

type userHandler struct {
	log      logger.Loggor
	producer sarama.SyncProducer
	cfg      *config.Config
	validate validator.Validate
	workerID int
}

func (h *userHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *userHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *userHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	for message := range claim.Messages() {
		h.log.Infof(
			"WORKER: %v, message at topic/partition/offset %v/%v/%v: %s = %s\n",
			h.workerID,
			message.Topic,
			message.Partition,
			message.Offset,
			string(message.Key),
			string(message.Value),
		)

		var user model.User
		if err := json.Unmarshal(message.Value, &user); err != nil {
			h.log.Errorf("json.Unmarshal", err)
			continue
		}
		user.Password = ""
		// Validate the user
		if err := h.validate.StructCtx(ctx, &user); err != nil {
			h.log.Errorf("validate.StructCtx", err)
			continue
		}

		// Publish the message to the appropriate Kafka topic
		if message.Topic == CreateUserTopic {
			h.log.Infof("created user: %v", user)
			_, _, err := h.producer.SendMessage(&sarama.ProducerMessage{
				Topic: CreateUserTopic,
				Value: sarama.StringEncoder(message.Value),
			})
			if err != nil {
				h.log.Errorf("Error sending message: %v", err)
			}
		} else if message.Topic == UpdateUserTopic {
			h.log.Debugf("updated user: %v", user)
			_, _, err := h.producer.SendMessage(&sarama.ProducerMessage{
				Topic: UpdateUserTopic,
				Value: sarama.StringEncoder(message.Value),
			})
			if err != nil {
				h.log.Errorf("Error sending message: %v", err)
			}
		}

		// Mark the message as processed
		session.MarkMessage(message, "")
	}
	return nil
}
