package rabbitmq

import (
	"Go-grpc/config"
	"Go-grpc/pkg/logger"
	"Go-grpc/pkg/rabbitmq"
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
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

type Publisher interface {
	CreateExchangeAndQueue(exchange, queueName, bindingKey string) (*amqp.Channel, error)
	Publish(ctx context.Context, exchange, routingKey, contentType string, header amqp.Table, body []byte) error
}

type UserPublisher struct {
	amqpConn *amqp.Connection
	cfg      *config.Config
	logger   logger.Loggor
}

func NewUserPublisher(cfg *config.Config, logger logger.Loggor) (*UserPublisher, error) {
	amqp, err := rabbitmq.NewRabbitMQConn(cfg)

	if err != nil {
		return nil, err
	}

	return &UserPublisher{amqpConn: amqp, cfg: cfg, logger: logger}, nil
}

func (p *UserPublisher) CreateExchangeAndQueue(exchange, queueName, bindingKey string) (*amqp.Channel, error) {
	amqpChan, err := p.amqpConn.Channel()

	if err != nil {
		return nil, errors.Wrap(err, "p.amqpConn.Channel")
	}

	p.logger.Infof("Declaring Exhange: %s", exchange)

	if err := amqpChan.ExchangeDeclare(
		exchange,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	); err != nil {
		return nil, errors.Wrap(err, "Error ch.ExchangeDeclaring")
	}

	queue, err := amqpChan.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueDeclaring")
	}

	p.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messageCount: %v, "+
		"consumerCount: %v, exchange: %v, exchange: %v, bindingKey: %v", queue.Name, queue.Messages, queue.Consumers, exchange, bindingKey)
	err = amqpChan.QueueBind(queue.Name, bindingKey, exchange, queueNoWait, nil)

	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueBinding")
	}

	return amqpChan, nil
}

func (p *UserPublisher) Publish(ctx context.Context, exchange, routingKey, contentType string, header amqp.Table, body []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UserPublisher.Publish")

	defer span.Finish()

	amqpChan, err := p.amqpConn.Channel()

	if err != nil {
		return errors.Wrap(err, "p.amqp.Channel")
	}

	defer amqpChan.Close()

	p.logger.Infof("Publishing messages exchange : %s, RoutingKey: %s", exchange, routingKey)

	if err := amqpChan.Publish(
		exchange,
		routingKey,
		publishMandatory,
		publishImmediate,
		amqp.Publishing{
			Headers:      header,
			ContentType:  contentType,
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.NewV4().String(),
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	); err != nil {
		errorPublisherMessages.Inc()
		return errors.Wrap(err, "ch.Publish")
	}

	successPublisherhMessages.Inc()
	return nil
}
