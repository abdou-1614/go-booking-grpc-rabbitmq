package rabbitmq

import (
	"Go-grpc/config"
	"Go-grpc/internal/user"
	"Go-grpc/pkg/logger"
	"Go-grpc/pkg/rabbitmq"
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type UserConsumer struct {
	amqpConn *amqp.Connection
	cfg      *config.Config
	logger   logger.Loggor
	userUC   user.UseCase
}

func NewUserConsumer(cfg *config.Config, logger logger.Loggor, user user.UseCase) *UserConsumer {
	return &UserConsumer{
		cfg:    cfg,
		logger: logger,
		userUC: user,
	}
}

func (p *UserConsumer) Dial() error {
	conn, err := rabbitmq.NewRabbitMQConn(p.cfg)

	if err != nil {
		return err
	}

	p.amqpConn = conn

	return nil
}

func (c *UserConsumer) CreateExchangeAndQueue(exchangeName, queueName, bindingKey string) (*amqp.Channel, error) {
	amqpChann, err := c.amqpConn.Channel()

	if err != nil {
		return nil, errors.Wrap(err, "c.amqpConn.Channel")
	}

	c.logger.Infof("Exchange Declaring : %s", exchangeName)

	if err := amqpChann.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeDurable,
		exchangeNoWait,
		nil,
	); err != nil {
		return nil, errors.Wrap(err, "Error amqpChann.ExchangeDeclaring")
	}

	queue, err := amqpChann.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Error amqpChann.QueueExchange")
	}

	c.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messagesCount: %v, "+
		"consumerCount: %v, exchange: %v, bindingKey: %v", queue.Name,
		queue.Messages,
		queue.Consumers,
		exchangeName,
		bindingKey)

	err = amqpChann.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		queueNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Error amqpChann.QueueBind")
	}

	err = amqpChann.Qos(
		prefetchCount,
		prefetchSize,
		prefetchGlobal,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Error  amqpChann.Qos")
	}

	return amqpChann, nil
}

func (c *UserConsumer) startConsumer(
	ctx context.Context,
	workerPoolSize int,
	queueName string,
	consumerTag string,
	worker func(ctx context.Context, wg *sync.WaitGroup, message <-chan amqp.Delivery),
) error {
	amqpChann, err := c.amqpConn.Channel()

	if err != nil {
		return errors.Wrap(err, "c.amqpConn.Channel")
	}

	deliveries, err := amqpChann.Consume(
		queueName,
		consumerTag,
		consumeAutoAck,
		consumeExclusive,
		consumeNoLocal,
		consumeNoWait,
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "amqpChann.Consume")
	}

	wg := &sync.WaitGroup{}

	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go worker(ctx, wg, deliveries)
	}

	amqpChannError := <-amqpChann.NotifyClose(make(chan *amqp.Error))

	c.logger.Errorf("amqpChann.NotifyClose : %s", amqpChannError)

	wg.Wait()

	return amqpChannError
}

func (c *UserConsumer) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		if err := c.startConsumer(
			ctx,
			AvatarWorker,
			AvatarQueueName,
			AvatarConsumerTag,
			c.imagesWorker,
		); err != nil {
			c.logger.Errorf("StartResizeConsumer: %v", err)
			cancel()
		}
	}()
}
