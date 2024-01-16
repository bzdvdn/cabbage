package cabbage

import (
	"errors"
	"fmt"
	"sync"
)

// CabbageClient provides API for sending cabbage tasks
type CabbageClient struct {
	broker         CabbageBroker
	taskLock       sync.RWMutex
	workers        map[string]*CabbageWorker
	publisher      *Publisher
	registredTasks map[string]*Task
}

// CabbageBroker is interface for cabbage broker db
type CabbageBroker interface {
	SendCabbageMessage(queueName string, cbMessage *CabbageMessage) error
	GetCabbageMessage(queueName string) (*CabbageMessage, error)
	EnableQueueForWorker(queueName string) error
	Close()
}

// NewCabbageClient create new CabbageClient
func NewCabbageClient(broker CabbageBroker) *CabbageClient {
	return &CabbageClient{
		broker:         broker,
		workers:        make(map[string]*CabbageWorker),
		registredTasks: make(map[string]*Task),
	}
}

// CreateWorker create cabbage worker for cunsume data
func (cc *CabbageClient) CreateWorker(queueName string, concurrency int) (*CabbageWorker, error) {
	err := cc.broker.EnableQueueForWorker(queueName)
	if err != nil {
		return nil, err
	}
	worker := newCabbageWorker(cc.broker, concurrency, queueName)
	_, ok := cc.workers[queueName]
	if ok {
		return nil, fmt.Errorf("worker for queue: %s exist", queueName)
	}
	cc.workers[queueName] = worker
	return worker, nil
}

// CreatePublisher create publisher for publish data to broker
func (cc *CabbageClient) CreatePublisher() *Publisher {
	publisher := newPublisher(cc.broker)
	cc.publisher = publisher
	return publisher
}

// Close connections
func (cc *CabbageClient) Close() {
	cc.broker.Close()
}

// RegisterTask register task for worker/publisher
func (cc *CabbageClient) RegisterTask(task *Task) error {
	cc.taskLock.Lock()
	if task.TProccesser != nil {
		worker, ok := cc.workers[task.QueueName]
		if ok {
			worker.RegisterTaskProcesser(task.Name, task.TProccesser)
		} else {
			cc.taskLock.Unlock()
			return errors.New("[!] try to register task proccesser, but not workers enabled")
		}
	}
	if task.WithPublish && cc.publisher != nil {
		cc.publisher.RegisterTask(task)
	}
	cc.taskLock.Unlock()
	return nil
}

// CreateScheduler create scheduler for schdule task
func (cc *CabbageClient) CreateScheduler() *Scheduler {
	scheduler := newScheduler(cc.broker)
	return scheduler
}
