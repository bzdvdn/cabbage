# Cabbage

Go Publisher/Worker/Scheduler for Distributed Task Queue

## Install

    go get -u github.com/bzdvdn/cabbage/cabbage

### Usage

RabbitMQ broker

```go
package main

import (
	"fmt"

	"github.com/bzdvdn/cabbage/cabbage"
)

func main() {
	broker, err := cabbage.NewRabbitMQBroker("amqp://<rq_connection>", 1)
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
    client := cabbage.NewCabbageClient(broker)
	defer client.Close()
}

```

Redis broker

```go
package main

import (
	"fmt"

	"github.com/bzdvdn/cabbage/cabbage"
)

func main() {
	broker, err := cabbage.NewRedisBroker("redis://<redis_connection>")
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
    client := cabbage.NewCabbageClient(broker)
	defer client.Close()
}

```

Create worker

```go
...

func main() {
    ...
    worker, err := client.CreateWorker("cabbageQueue", 4)
    if err != nil {
		fmt.Printf("error %v", err)
		return
	}
    ...
    // Start worker
    err = worker.StartWorker()
	if err != nil {
		fmt.Println(err)
		return
	}
    ...
    // Start worker with context
    err = worker.StartWorkerWithContext(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
}

```

Create Publisher

```go
...

func main() {
    ...
    publisher := client.CreatePublisher()
    ...
}

```

Create Scheduler

```go
...

func main() {
    ...
    scheduler := client.CreateScheduler()
    ...
    // start scheduler
    <-scheduler.Start()
}

```

Create tasks

```go
type TaskData struct {
	Test  int      `json:"test"`
	Value string   `json:"value"`
	Array []string `json:"array"`
}

type TestService struct {
}

func (t *TestService) ProccessTask(ctx context.Context, body []byte, ID string) error {
	var td TaskData
	err := json.Unmarshal(body, &td)
	if err != nil {
		return err
	}
	fmt.Println(t)
	fmt.Println("task done!")
	return nil
}


func main() {
    ...
    task, err := cabbage.NewTask("cabbageQueue", "TestTask", &TestService{}, false)
    // add task to publisher
    task, err := cabbage.NewTask("cabbageQueue", "TestTask", &TestService{}, true)
    if err != nil {
        fmt.Println(err)
        return
    }
	client.RegisterTask(task)
    ...
}

```

Create ScheduleTask

```go
type SchData struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
}

func (s *SchData) ToPublish() ([]byte, error) {
	js, err := json.Marshal(s)
	if err != nil {
		return []byte(""), err
	}
	return js, nil
}

func schedule1() (tpublisher cabbage.TaskPublisher) {
	d := SchData{ID: "123", SiteID: "321"}
	return &d
}

func schedule2() (tpublisher cabbage.TaskPublisher) {
	d := SchData{ID: "222", SiteID: "2222"}
	return &d
}

func main() {
    ...
    every5min, _ := cabbage.EntryEveryMinute(5)
	tasks := []*cabbage.ScheduleTask{
		{
			Name:      "shdtask",
			QueueName: "cabbageQueue",
			Func:      schedule1,
			Entries:   cabbage.Entries{&cabbage.Entry{Schedule: "* * * * *"}},
		}, {
			Name:      "shdtask2",
			QueueName: "cabbageQueue",
			Func:      schedule2,
			Entries:   cabbage.Entries{every5min},
		}
    }
    scheduler.AddScheduleTasks(tasks)
    ...
    // <-scheduler.Start()
}

```
