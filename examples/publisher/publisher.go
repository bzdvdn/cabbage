package main

import (
	"encoding/json"
	"fmt"

	"github.com/bzdvdn/cabbage/cabbage"
)

type TestData struct {
	Test  int      `json:"test"`
	Value string   `json:"value"`
	Array []string `json:"array"`
}

func (t TestData) ToPublish() ([]byte, error) {
	js, err := json.Marshal(t)
	if err != nil {
		return []byte(""), err
	}
	return js, nil
}

func main() {
	broker, err := cabbage.NewRabbitMQBroker("amqp://admin:Gfhjkm123@localhost:5672/", 1)
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
	client := cabbage.NewCabbageClient(broker)
	defer client.Close()
	publisher := client.CreatePublisher()
	task := cabbage.Task{QueueName: "cabbageQueue", Name: "TestTask", WithPublish: true}
	client.RegisterTask(&task)
	ts1 := TestData{Test: 1, Value: "1", Array: []string{"1"}}
	err = publisher.PublishTask("TestTask", &ts1)
	fmt.Println(err)
}
