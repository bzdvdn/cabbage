package cabbage

import (
	"context"
	"encoding/json"
	"testing"
)

type MockCabbageBroker struct{}

func (m *MockCabbageBroker) SendCabbageMessage(queueName string, cbMessage *CabbageMessage) error {
	return nil
}

func (m *MockCabbageBroker) EnableQueueForWorker(queueName string) error {
	return nil
}

func (m *MockCabbageBroker) GetCabbageMessage(queueName string) (*CabbageMessage, error) {
	return cbMessage, nil
}

func (m *MockCabbageBroker) Close() {}

type TestService struct {
	Test  int    `json:"test"`
	Value string `json:"value"`
}

func (t TestService) ProccessTask(ctx context.Context, body []byte, ID string) error {
	err := json.Unmarshal(body, &t)
	if err != nil {
		return err
	}
	return nil
}

var (
	broker      = &MockCabbageBroker{}
	tProccesser = TestService{}
	testTask    = &Task{
		Name:        taskName,
		QueueName:   queueName,
		TProccesser: tProccesser,
	}
)

func TestCreateWorkerAndRegisterTask(t *testing.T) {
	client := NewCabbageClient(broker)
	worker, err := client.CreateWorker(queueName, 1)
	if err != nil {
		t.Log("Cant create worker")
		t.Fail()
	}
	if len(client.workers) > 1 {
		t.Log("Invalid workers len")
		t.Fail()
	}
	if client.workers[queueName] != worker {
		t.Logf("Invalid worker for queue: %s", queueName)
		t.Fail()
	}
	_, err = client.CreateWorker(queueName, 1)
	if err == nil {
		t.Log("failed check same worker for queue")
		t.Fail()
	}
	err = client.RegisterTask(testTask)
	if err != nil {
		t.Logf("cant register task: %v", err)
		t.Fail()
	}
}
