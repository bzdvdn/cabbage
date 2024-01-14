package cabbage

import (
	"testing"
	"time"
)

func TestMesssageSendAndConsumeFromRabbitMQ(t *testing.T) {
	broker := testNewRQBroker(t)
	defer broker.Close()
	err := broker.EnableQueueForWorker(queueName)
	if err != nil {
		t.Fatalf("1, %v", err)
	}
	err = broker.SendCabbageMessage(queueName, cbMessage)
	if err != nil {
		t.Fatalf("cant send cb message to rabbitmq, %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	msg, err := broker.GetCabbageMessage(queueName)
	if err != nil {
		t.Fatalf("cant get cb message from rabbitmq, %v", err)
	}
	if msg.ID != cbMessage.ID {
		t.Log("Invalid ids in rabbitmq")
		t.Fail()
	}
	if string(msg.Body) != string(cbMessage.Body) {
		t.Log("Invalid bodies in rabbitmq")
		t.Fail()
	}
}
