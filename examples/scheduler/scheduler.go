package main

import (
	"encoding/json"
	"fmt"

	"github.com/bzdvdn/cabbage/cabbage"
)

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
	d := SchData{ID: "uuuu1", SiteID: "id1"}
	return &d
}
func schedule2() (tpublisher cabbage.TaskPublisher) {
	d := SchData{ID: "uuuu2", SiteID: "id2"}
	return &d
}
func schedule3() (tpublisher cabbage.TaskPublisher) {
	d := SchData{ID: "uuuu3", SiteID: "id3"}
	return &d
}

func main() {
	every5min, _ := cabbage.EntryEveryMinute(5)
	every, _ := cabbage.EntryEvery([]int{}, []int{2, 5, 10}, []int{}, []int{}, []int{0}) // every 2,5,10 hours at Sunday
	tasks := []*cabbage.ScheduleTask{
		{
			Name:      "shdtask",
			QueueName: "scheduleQueue1",
			Func:      schedule1,
			Entries:   cabbage.Entries{&cabbage.Entry{Schedule: "* * * * *"}},
		}, {
			Name:      "shdtask2",
			QueueName: "scheduleQueue2",
			Func:      schedule2,
			Entries:   cabbage.Entries{every5min},
		}, {
			Name:      "shdtask3",
			QueueName: "SundayQueue",
			Func:      schedule3,
			Entries:   cabbage.Entries{every},
		},
	}

	broker, err := cabbage.NewRedisBroker("redis://:password@localhost:6379/")
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
	if err != nil {
		fmt.Printf("error %v", err)
		return
	}
	client := cabbage.NewCabbageClient(broker)
	defer client.Close()
	scheduler := client.CreateScheduler()
	scheduler.AddScheduleTasks(tasks)
	<-scheduler.Start()
}
