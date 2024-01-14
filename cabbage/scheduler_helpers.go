package cabbage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// EntryEveryMinute create *Entry with every minute cron
func EntryEveryMinute(minute int) (*Entry, error) {
	if minute < 0 {
		return nil, errors.New("minute cant be < 0")
	} else if minute > 59 {
		return nil, errors.New("minute cant be > 59")
	}
	return &Entry{Schedule: fmt.Sprintf("*/%d * * * *", minute)}, nil
}

// EntryEveryHour create *Entry with every hour cron
func EntryEveryHour(hour int) (*Entry, error) {
	if hour < 0 {
		return nil, errors.New("hour cant be < 0")
	} else if hour > 23 {
		return nil, errors.New("hour cant be > 23")
	}
	return &Entry{Schedule: fmt.Sprintf("0 */%d * * *", hour)}, nil
}

// EntryEveryHourAtMinute create *Entry with every hour and minute cron
func EntryEveryHourAtMinute(minute int, hour int) (*Entry, error) {
	if hour < 0 {
		return nil, errors.New("hour cant be < 0")
	} else if hour > 23 {
		return nil, errors.New("hour cant be > 23")
	}
	if minute < 0 {
		return nil, errors.New("minute cant be < 0")
	} else if minute > 59 {
		return nil, errors.New("minute cant be > 59")
	}
	return &Entry{Schedule: fmt.Sprintf("%d */%d * * *", minute, hour)}, nil
}

// EntryEvery generate *Entry by minutes, hours, days, months, week days slices
func EntryEvery(minutes []int, hours []int, days []int, months []int, weekDays []int) (*Entry, error) {
	minute_string := "*"
	hour_string := "*"
	day_string := "*"
	month_string := "*"
	weekday_string := "*"
	if len(minutes) > 0 {
		minute_string = intSliceToString(minutes)
	}
	if len(hours) > 0 {
		hour_string = intSliceToString(hours)
	}
	if len(days) > 0 {
		day_string = intSliceToString(days)
	}
	if len(months) > 0 {
		month_string = intSliceToString(months)
	}
	if len(weekDays) > 0 {
		weekday_string = intSliceToString(weekDays)
	}
	schedule := fmt.Sprintf("%s %s %s %s %s", minute_string, hour_string, day_string, month_string, weekday_string)
	_, err := parseSchedule(schedule)
	return &Entry{Schedule: schedule}, err
}

func intSliceToString(sl []int) string {
	strSlice := make([]string, len(sl))
	for i, v := range sl {
		strSlice[i] = strconv.Itoa(v)
	}

	return strings.Join(strSlice, ",")
}
