package rabbitmq

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/streadway/amqp"
)

func (c *UserConsumer) imagesWorker(ctx context.Context, wg *sync.WaitGroup, messages <-chan amqp.Delivery) {
	defer wg.Done()

	for delivery := range messages {
		span, _ := opentracing.StartSpanFromContext(ctx, "ImageConsumer.resizeWorker")

		c.logger.Infof("processDeliveries deliveryTag% v", delivery.DeliveryTag)

		incomingMessages.Inc()
		span.Finish()
	}

	c.logger.Info("Deliveries channel closed")
}
