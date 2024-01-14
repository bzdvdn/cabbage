package cabbage

import "testing"

func TestMesssageSendAndConsumeFromRedis(t *testing.T) {
	broker := testNewRedisBroker(t)
	defer broker.Close()
	err := broker.SendCabbageMessage(queueName, cbMessage)
	if err != nil {
		t.Fatalf("cant send cb message to redis, %v", err)
	}
	msg, err := broker.GetCabbageMessage(queueName)
	if err != nil {
		t.Fatalf("cant get cb message from redis, %v", err)
	}
	if msg.ID != cbMessage.ID {
		t.Log("Invalid ids in redis")
		t.Fail()
	}
	if string(msg.Body) != string(cbMessage.Body) {
		t.Log("Invalid bodies in redis")
		t.Fail()
	}
}
