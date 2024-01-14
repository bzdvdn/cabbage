package cabbage

import (
	"context"
	"errors"
)

// TaskProccesser interface for proccess consume message
type TaskProccesser interface {
	ProccessTask(ctx context.Context, body []byte, ID string) error
}

// TaskPublisher interface for publish message
type TaskPublisher interface {
	ToPublish() ([]byte, error)
}

// Task cabbage task struct
type Task struct {
	Name        string
	QueueName   string
	TProccesser TaskProccesser
	WithPublish bool
}

// NewTask construct cabbage Task
func NewTask(name string, queueName string, tproccesser TaskProccesser, withPublish bool) (*Task, error) {
	if name == "" {
		return nil, errors.New("task name cant be empty")
	}
	if queueName == "" {
		return nil, errors.New("queueName cant be empty")
	}
	return &Task{
		Name:        name,
		QueueName:   queueName,
		TProccesser: tproccesser,
		WithPublish: withPublish}, nil
}

// TaskProccessersRoutes router for workers consume
type TaskProccessersRoutes struct {
	Routes map[string]TaskProccesser
}

// NewTaskProccessersRoutes construct TaskProccessersRoutes
func NewTaskProccessersRoutes(routes map[string]TaskProccesser) *TaskProccessersRoutes {
	return &TaskProccessersRoutes{Routes: routes}
}
