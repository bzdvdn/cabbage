package cabbage

import (
	"os"
	"testing"
)

var (
	queueName = "unittestQueue"
	taskName  = "testTask"
	body      = []byte(`{"id": "unittestId"}`)
	cbMessage = newCabbageMessage(taskName, body)
)

func testNewRQBroker(t *testing.T) *RabbitMQBroker {
	url := os.Getenv("RQ_HOST")
	broker, err := NewRabbitMQBroker(url, 1)
	if err != nil {
		t.Fatalf("cant connect to RabbitMQ, %v", err)
	}
	return broker
}

func testNewRedisBroker(t *testing.T) *RedisBroker {
	url := os.Getenv("REDIS_HOST")
	broker, err := NewRedisBroker(url)
	if err != nil {
		t.Fatalf("cant connect to Redis, %v", err)
	}
	return broker
}
