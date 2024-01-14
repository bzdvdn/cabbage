package cabbage

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestParseScheduleError(t *testing.T) {
	var scheduleErrorTestSlice = []string{
		"* * * * * *",
		"0-100 * * * *",
		"* -1-23 * * *",
		"* * -1-22 * *",
		"* * 1,2,32 * *",
		"* * 1-40/2 * *",
		"* * qw/2 * *",
		"* * * -1-12 *",
		"* * * 0-12 *",
		"* * * 1-13 *",
		"* * * * -1,5",
		"* * * * 0,8",
		"* * * * j,0",
		"1 2 3 4 5 6",
		"* 1,2/10 * * *",
		"* * 1,2,3,1-15/10 * *",
		"q w e r t",
	}

	for _, s := range scheduleErrorTestSlice {
		if _, err := parseSchedule(s); err == nil {
			t.Error(s, "should be error", err)
		}
	}
}

func TestParseSchedule(t *testing.T) {
	var schTest = []struct {
		schedule string
		count    [5]int
	}{
		{"* * * * *", [5]int{60, 24, 31, 12, 7}},
		{"*/2 * * * *", [5]int{30, 24, 31, 12, 7}},
		{"*/10 * * * *", [5]int{6, 24, 31, 12, 7}},
		{"* * * * */2", [5]int{60, 24, 0, 12, 4}},
		{"5,8,9 */2 2,3 * */2", [5]int{3, 12, 2, 12, 4}},
		{"* 5-11 2-30/2 * *", [5]int{60, 7, 15, 12, 0}},
		{"1,2,5-8 * * */3 *", [5]int{6, 24, 31, 4, 7}},
	}

	for _, sch := range schTest {
		j, err := parseSchedule(sch.schedule)
		if err != nil {
			t.Error(err)
		}

		if len(j.min) != sch.count[0] {
			t.Error(sch.schedule, "min count expected to be", sch.count[0], "result", len(j.min), j.min)
		}

		if len(j.hour) != sch.count[1] {
			t.Error(sch.schedule, "hour count expected to be", sch.count[1], "result", len(j.hour), j.hour)
		}

		if len(j.day) != sch.count[2] {
			t.Error(sch.schedule, "day count expected to be", sch.count[2], "result", len(j.day), j.day)
		}

		if len(j.month) != sch.count[3] {
			t.Error(sch.schedule, "month count expected to be", sch.count[3], "result", len(j.month), j.month)
		}

		if len(j.weekDay) != sch.count[4] {
			t.Error(sch.schedule, "weekDay count expected to be", sch.count[4], "result", len(j.weekDay), j.weekDay)
		}
	}
}

type testSchData struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
}

func (s *testSchData) ToPublish() ([]byte, error) {
	js, err := json.Marshal(s)
	if err != nil {
		return []byte(""), err
	}
	return js, nil
}
func schedule1() (tpublisher TaskPublisher) {
	d := testSchData{ID: "hsfhsjghjs", SiteID: "mnbghs"}
	return &d
}

func TestScheduler(t *testing.T) {
	testShTasks := []*ScheduleTask{
		{
			Name:      "shdtask",
			QueueName: "SHTEST_RQ",
			Func:      schedule1,
			Entries:   Entries{&Entry{Schedule: "* * * * *"}},
		},
	}
	client := NewCabbageClient(testbroker)
	defer client.Close()
	scheduler := client.CreateScheduler()
	err := scheduler.AddScheduleTasks(testShTasks)
	if err != nil {
		t.Fatalf("cant add SheduleTasks, %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case <-scheduler.Start():
			continue
		}
	}
}
