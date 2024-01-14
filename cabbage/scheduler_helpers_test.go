package cabbage

import (
	"testing"
)

func TestEntryEveryMinute(t *testing.T) {
	_, err := EntryEveryMinute(-1)
	if err == nil {
		t.Log("invalid error handle, must handle minute < 0")
		t.Fail()
	}
	entry, err := EntryEveryMinute(5)
	if err != nil {
		t.Logf("invalid error %v \n", err)
		t.Fail()
	}
	if entry.Schedule != "*/5 * * * *" {
		t.Log("Invalid schedule in EntryEveryMinute, must be */5 * * * *")
		t.Fail()
	}
}

func TestEntryEveryHour(t *testing.T) {
	_, err := EntryEveryHour(-1)
	if err == nil {
		t.Log("invalid error handle, must handle hour < 0")
		t.Fail()
	}
	entry, err := EntryEveryHour(5)
	if err != nil {
		t.Logf("invalid error %v \n", err)
		t.Fail()
	}
	if entry.Schedule != "0 */5 * * *" {
		t.Log("Invalid schedule in EntryEveryHour, must be 0 */5 * * *")
		t.Fail()
	}
}

func TestEntryEveryHourAtMinute(t *testing.T) {
	_, err := EntryEveryHourAtMinute(-1, 2)
	if err == nil {
		t.Log("invalid error handle, must handle minute < 0")
		t.Fail()
	}
	_, err = EntryEveryHourAtMinute(25, -5)
	if err == nil {
		t.Log("invalid error handle, must minute hour < 0")
		t.Fail()
	}
	entry, err := EntryEveryHourAtMinute(25, 4)
	if err != nil {
		t.Logf("invalid error %v \n", err)
		t.Fail()
	}
	if entry.Schedule != "25 */4 * * *" {
		t.Log("Invalid schedule in EntryEveryHourAtMinute, must be 25, */4 * * *")
		t.Fail()
	}
}

func TestEntryEvery(t *testing.T) {
	_, err := EntryEvery([]int{-1, 2}, []int{}, []int{}, []int{}, []int{})
	if err == nil {
		t.Log("minute validation failed in EntryEvery")
		t.Fail()
	}
	_, err = EntryEvery([]int{}, []int{-1, 2}, []int{}, []int{}, []int{})
	if err == nil {
		t.Log("hour validation failed in EntryEvery")
		t.Fail()
	}
	_, err = EntryEvery([]int{}, []int{}, []int{-1, 2}, []int{}, []int{})
	if err == nil {
		t.Log("day validation failed in EntryEvery")
		t.Fail()
	}
	_, err = EntryEvery([]int{}, []int{}, []int{}, []int{32}, []int{})
	if err == nil {
		t.Log("month validation failed in EntryEvery")
		t.Fail()
	}
	_, err = EntryEvery([]int{}, []int{}, []int{}, []int{}, []int{22})
	if err == nil {
		t.Log("weekday validation failed in EntryEvery")
		t.Fail()
	}
	entry, err := EntryEvery([]int{}, []int{22, 4, 9, 14, 18}, []int{}, []int{}, []int{})
	if err != nil {
		t.Logf("invalid validation, %v \n", err)
		t.Fail()
	}
	if entry.Schedule != "* 22,4,9,14,18 * * *" {
		t.Log("Invalid schedule hours in EntryEvery")
		t.Fail()
	}
	entry, err = EntryEvery([]int{22, 4, 9, 14, 18}, []int{}, []int{}, []int{}, []int{})
	if err != nil {
		t.Logf("invalid validation, %v \n", err)
		t.Fail()
	}
	if entry.Schedule != "22,4,9,14,18 * * * *" {
		t.Log("Invalid schedule hours in EntryEvery")
		t.Fail()
	}
}
