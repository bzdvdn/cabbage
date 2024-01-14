package cabbage

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type ConsumingChannels map[string]<-chan amqp.Delivery

// RabbitMQBroker implement rabbtimq broker
type RabbitMQBroker struct {
	connection        *amqp.Connection
	channel           *amqp.Channel
	consumingChannels ConsumingChannels
	rate              int
}

// RabbitMQQueue queue for rabbitmq
type RabbitMQQueue struct {
	Name       string
	Durable    bool
	AutoDelete bool
}

// newRabbitMQQueue construct RabbitMQQueue
func newRabbitMQQueue(name string) *RabbitMQQueue {
	return &RabbitMQQueue{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
	}
}

// NewRabbitMQBroker constructor for RabbitmqBroker
func NewRabbitMQBroker(url string, rate int) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("RQ.Connection %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("conn.channel %w", err)
	}
	broker := &RabbitMQBroker{
		connection:        conn,
		channel:           ch,
		rate:              rate,
		consumingChannels: make(map[string]<-chan amqp.Delivery),
	}
	if err := broker.channel.Qos(broker.rate, 0, false); err != nil {
		log.Println("Cant Qos RQ channel")
		return nil, err
	}
	return broker, nil
}

// createExchangeName generate exchange name
func (b *RabbitMQBroker) createExchangeName(queueName string) string {
	return fmt.Sprintf("%s_cabbage_exchange", queueName)
}

// createQueue declares RabbitMQQueue with stored configuration
func (b *RabbitMQBroker) createQueue(queueName string) error {
	q := newRabbitMQQueue(queueName)
	exchangeName := b.createExchangeName(queueName)
	err := b.channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	_, err = b.channel.QueueDeclare(
		q.Name,
		q.Durable,
		q.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = b.channel.QueueBind(q.Name, q.Name, exchangeName, false, nil)
	return err
}

// Close close broker connections
func (b *RabbitMQBroker) Close() {
	b.connection.Close()
	b.channel.Close()
}

// EnableQueueForWorker create queue and start consume from rabbitmq broker
func (b *RabbitMQBroker) EnableQueueForWorker(queueName string) error {
	if err := b.createQueue(queueName); err != nil {
		return err
	}
	if err := b.startConsumingChannel(queueName); err != nil {
		return err
	}
	return nil
}

// startConsumingChannel spawns receiving channel on AMQP queue
func (b *RabbitMQBroker) startConsumingChannel(queueName string) error {
	channel, err := b.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	b.consumingChannels[queueName] = channel
	return nil
}

// GetCabbageMessage get cabbage message from broker
func (b *RabbitMQBroker) GetCabbageMessage(queueName string) (*CabbageMessage, error) {
	select {
	case delivery := <-b.consumingChannels[queueName]:
		deliveryAck(delivery)
		messageId := delivery.MessageId
		if messageId == "" {
			messageId = "<EMPTY>"
		}
		cMessage := CabbageMessage{
			ID:        delivery.Headers["id"].(string),
			Body:      delivery.Body,
			MessageId: messageId,
			Timestamp: delivery.Timestamp,
			TaskName:  delivery.Headers["taskName"].(string),
		}

		return &cMessage, nil
	default:
		return nil, fmt.Errorf("consuming channel is empty")
	}
}

// SendCabbageMessage send cabbage message to broker
func (b *RabbitMQBroker) SendCabbageMessage(queueName string, cbMessage *CabbageMessage) error {
	if err := b.createQueue(queueName); err != nil {
		return err
	}
	publishMessage := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Headers:      amqp.Table{"id": cbMessage.ID, "taskName": cbMessage.TaskName},
		ContentType:  "application/json",
		Body:         cbMessage.Body,
		Timestamp:    cbMessage.Timestamp,
	}
	return b.channel.Publish(
		b.createExchangeName(queueName),
		queueName,
		false,
		false,
		publishMessage,
	)
}

// deliveryAck acknowledges delivery message with retries on error
func deliveryAck(delivery amqp.Delivery) {
	var err error
	for retryCount := 3; retryCount > 0; retryCount-- {
		if err = delivery.Ack(false); err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("rabbitmq_broker: failed to acknowledge result message %+v: %+v", delivery.MessageId, err)
	}
}
