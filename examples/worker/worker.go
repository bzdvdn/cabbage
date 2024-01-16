package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bzdvdn/cabbage/cabbage"
)

type TestData struct {
	Test  int      `json:"test"`
	Value string   `json:"value"`
	Array []string `json:"array"`
}

type TestService struct {
}

func (t TestService) ProccessTask(ctx context.Context, body []byte, ID string) error {
	var td TestData
	err := json.Unmarshal(body, &td)
	if err != nil {
		return err
	}
	fmt.Println("Start task")
	fmt.Println(t)
	return nil
}

func NewTestService() *TestService {
	return &TestService{}
}

func main() {
	broker, err := cabbage.NewRabbitMQBroker("amqp://admin:admin@localhost:5672/", 1)
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
	client := cabbage.NewCabbageClient(broker)
	worker, _ := client.CreateWorker("cabbageQueue", 4)
	task := cabbage.Task{QueueName: "cabbageQueue", Name: "TestTask", TProccesser: NewTestService()}
	client.RegisterTask(&task)
	defer client.Close()
	fmt.Println("Successfully Connected to our RabbitMQ Instance")
	fmt.Println(" [*] - Waiting for messages")
	ctx, cancel := context.WithCancel(context.Background())
	err = worker.StartWorkerWithContext(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Wait for OS exit signal
	<-exit
	cancel()
	log.Println("Got exit signal")
	worker.StopWorker()
}
