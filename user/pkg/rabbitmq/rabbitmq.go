package rabbitmq

import (
	"Go-grpc/config"
	"fmt"

	"github.com/streadway/amqp"
)

func NewRabbitMQConn(cfg *config.Config) (*amqp.Connection, error) {
	connAddress := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	return amqp.Dial(connAddress)
}
