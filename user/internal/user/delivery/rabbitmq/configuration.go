package rabbitmq

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	exchangeKind       = "direct"
	exchangeDurable    = true
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false

	queueDurable    = true
	queueAutoDelete = false
	queueExclusive  = false
	queueNoWait     = false

	publishMandatory = false
	publishImmediate = false

	prefetchCount  = 1
	prefetchSize   = 0
	prefetchGlobal = false

	consumeAutoAck   = false
	consumeExclusive = false
	consumeNoLocal   = false
	consumeNoWait    = false

	UserExchange = "users"

	AvatarQueueName   = "avatar_queue"
	AvatarConsumerTag = "user_avatar_consumer"
	AvatarWorker      = 5
	AvatarsBindingKey = "update_avatar_key"
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rabbitmq_images_incoming_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})

	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rabbitmq_images_success_messages_total",
		Help: "The total numver of sucess incoming Rabbitmq messages",
	})

	errorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rabbitmq_images_error_messages_total",
		Help: "The total number of error incoming success RabbitMQ messages",
	})
)
