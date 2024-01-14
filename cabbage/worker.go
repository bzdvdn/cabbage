package cabbage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// CabbageWorker represents distributed task worker
type CabbageWorker struct {
	broker                   CabbageBroker
	numWorkers               int
	registeredTaskProcessers *TaskProccessersRoutes
	taskLock                 sync.RWMutex
	cancel                   context.CancelFunc
	workWG                   sync.WaitGroup
	rateLimitPeriod          time.Duration
	queueName                string
}

// newCabbageWorker construct CabbageWorker
func newCabbageWorker(broker CabbageBroker, numWorkers int, queueName string) *CabbageWorker {
	worker := &CabbageWorker{
		broker:          broker,
		numWorkers:      numWorkers,
		rateLimitPeriod: 100 * time.Millisecond,
		queueName:       queueName,
	}
	return worker
}

// RegisterTaskProccessersRoutes register task proccessers routes
func (w *CabbageWorker) RegisterTaskProccessersRoutes(tr *TaskProccessersRoutes) {
	w.taskLock.Lock()
	w.registeredTaskProcessers = tr
	w.taskLock.Unlock()
}

// RegisterTaskProcesser register task proccesser
func (w *CabbageWorker) RegisterTaskProcesser(name string, tproccesser TaskProccesser) {
	w.taskLock.Lock()
	if w.registeredTaskProcessers == nil {
		routes := make(map[string]TaskProccesser)
		routes[name] = tproccesser
		w.registeredTaskProcessers = NewTaskProccessersRoutes(routes)
	} else {
		w.registeredTaskProcessers.Routes[name] = tproccesser
	}
	w.taskLock.Unlock()
}

// StartWorkerWithContext start cabbage worker with context
func (w *CabbageWorker) StartWorkerWithContext(ctx context.Context) error {
	if w.registeredTaskProcessers == nil {
		return errors.New("not registred tasks")
	}
	var wctx context.Context
	wctx, w.cancel = context.WithCancel(ctx)
	w.workWG.Add(w.numWorkers)
	for i := 0; i < w.numWorkers; i++ {
		go func(workerID int) {
			log.Printf("[*] Start Cabbage Worker for Queue: %s, ID: %d \n", w.queueName, workerID)
			defer w.workWG.Done()
			ticker := time.NewTicker(w.rateLimitPeriod)
			for {
				select {
				case <-wctx.Done():
					log.Printf("[*] Finish Cabbage Worker for Queue: %s, ID: %d \n", w.queueName, workerID)
					return
				case <-ticker.C:
					// get task
					cbMessage, err := w.broker.GetCabbageMessage(w.queueName)
					if err != nil || cbMessage == nil {
						continue
					}
					// get task proccesser
					log.Printf("[*] Queue: %s, worker: %d, GET message\n", w.queueName, workerID)
					tp, err := w.getTaskProcesser(cbMessage.TaskName)
					if err != nil {
						log.Printf("[!] Queue: %s, worker: %d, cant get task proccesser for taskName %s, id %s: %+v", w.queueName, workerID, cbMessage.TaskName, cbMessage.ID, err)
						continue
					}
					// process task request
					err = w.runTask(ctx, tp, cbMessage)
					if err != nil {
						log.Printf("[!] Queue: %s, worker: %d,failed to run task message %s: %+v", w.queueName, workerID, cbMessage.ID, err)
						continue
					}
				}
			}
		}(i)
	}
	return nil
}

// StartWorker start cabbage worker
func (w *CabbageWorker) StartWorker() error {
	return w.StartWorkerWithContext(context.Background())
}

// getTaskProcesser get task proccesser
func (w *CabbageWorker) getTaskProcesser(taskName string) (TaskProccesser, error) {
	w.taskLock.RLock()
	task, ok := w.registeredTaskProcessers.Routes[taskName]
	if !ok {
		w.taskLock.RUnlock()
		return nil, fmt.Errorf("not registred task %s", taskName)
	}
	w.taskLock.RUnlock()
	return task, nil
}

// runTask run task from task proccesser interface
func (w *CabbageWorker) runTask(ctx context.Context, tp TaskProccesser, cbMessage *CabbageMessage) error {
	err := tp.ProccessTask(ctx, cbMessage.Body, cbMessage.ID)
	return err
}

// StopWorker stops cabbage workers
func (w *CabbageWorker) StopWorker() {
	w.cancel()
	w.workWG.Wait()
}

// StopWait waits for cabbage workers to terminate
func (w *CabbageWorker) StopWait() {
	w.workWG.Wait()
}
