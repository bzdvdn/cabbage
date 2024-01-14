package cabbage

import (
	"errors"
	"sync"
)

// Publisher cabbage task publisher
type Publisher struct {
	broker         CabbageBroker
	taskLock       sync.RWMutex
	registredTasks map[string]*Task
}

// newPublisher create Publisher
func newPublisher(broker CabbageBroker) *Publisher {
	return &Publisher{broker: broker, registredTasks: make(map[string]*Task)}
}

// PublishTask publish task to broker
func (p *Publisher) PublishTask(taskName string, tpublisher TaskPublisher) error {
	task, ok := p.registredTasks[taskName]
	if !ok {
		return errors.New("missing task")
	}
	body, err := tpublisher.ToPublish()
	if err != nil {
		return err
	}
	cbMessage := newCabbageMessage(taskName, body)
	if err := p.broker.SendCabbageMessage(task.QueueName, cbMessage); err != nil {
		return err
	}
	return nil
}

// RegisterTask register task in publisher
func (p *Publisher) RegisterTask(task *Task) {
	p.taskLock.Lock()
	p.registredTasks[task.Name] = task
	p.taskLock.Unlock()
}
