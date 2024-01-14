package cabbage

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// regexps for parsing schedule string
var (
	matchSpaces = regexp.MustCompile(`\s+`)
	matchN      = regexp.MustCompile(`(.*)/(\d+)`)
	matchRange  = regexp.MustCompile(`^(\d+)-(\d+)$`)
)

// Scheduler cabbage jobs scheduler
type Scheduler struct {
	broker    CabbageBroker
	ticker    *time.Ticker
	jobs      []*job
	shTasks   []*ScheduleTask
	publisher *Publisher
	stopped   chan bool
	sync.RWMutex
}

// tick is individual tick that occures each minute
type tick struct {
	min     int
	hour    int
	day     int
	month   int
	weekDay int
}

// Entry schedule cron syntax
type Entry struct {
	Schedule string
}

type Entries []*Entry

// ScheduleTask task scheduler
type ScheduleTask struct {
	Name      string
	QueueName string
	Func      func() (tpublisher TaskPublisher)
	Entries   Entries // cron schedule slice of Entry
}

// newScheduler construct scheduler
func newScheduler(broker CabbageBroker) *Scheduler {
	publisher := newPublisher(broker)
	scheduler := &Scheduler{
		broker:    broker,
		ticker:    time.NewTicker(time.Minute),
		publisher: publisher,
	}
	return scheduler
}

// job declare
type job struct {
	taskName  string
	publisher *Publisher
	min       map[int]struct{}
	hour      map[int]struct{}
	day       map[int]struct{}
	month     map[int]struct{}
	weekDay   map[int]struct{}
	fn        func() (tpublisher TaskPublisher)
	sync.RWMutex
}

// runPending jobs
func (shd *Scheduler) runPending(t time.Time) {
	tick := getTick(t)
	shd.RLock()
	defer shd.RUnlock()
	for _, j := range shd.jobs {
		if j.tick(tick) {
			go j.run()
		}
	}
}

// Start scheduler
func (shd *Scheduler) Start() chan bool {
	log.Println("[*] Start Cabbage Scheduler")
	shd.stopped = make(chan bool, 1)
	go func() {
		for {
			select {
			case t := <-shd.ticker.C:
				shd.runPending(t)
			case <-shd.stopped:
				shd.ticker.Stop()
				return
			}
		}
	}()
	return shd.stopped
}

// Shutdown scheduler
func (shd *Scheduler) Shutdown() {
	shd.stopped <- true
}

// AddScheduleTask add ScheduleTask to scheduler
func (shd *Scheduler) AddScheduleTask(shTask *ScheduleTask) error {
	shd.Lock()
	task, err := NewTask(shTask.Name, shTask.QueueName, nil, true)
	if err != nil {
		return err
	}
	shd.publisher.RegisterTask(task)
	if len(shTask.Entries) < 1 {
		return errors.New("[!] entries cant be empty")
	}
	for _, s := range shTask.Entries {
		if err := shd.addJob(s.Schedule, shTask.Name, shTask.Func); err != nil {
			return err
		}
	}
	shd.shTasks = append(shd.shTasks, shTask)
	shd.Unlock()
	return nil
}

// AddScheduleTasks add slice of ScheduleTasks to scheduler
func (shd *Scheduler) AddScheduleTasks(shTasks []*ScheduleTask) error {
	for _, shtask := range shTasks {
		if err := shd.AddScheduleTask(shtask); err != nil {
			return err
		}
	}
	return nil
}

// addJob add job to scheduler
func (shd *Scheduler) addJob(schedule string, taskName string, fn func() (tpublisher TaskPublisher)) error {
	j, err := parseSchedule(schedule)
	if err != nil {
		shd.Unlock()
		return err
	}
	j.publisher = shd.publisher
	j.taskName = taskName
	j.fn = fn
	shd.jobs = append(shd.jobs, j)
	return nil
}

// tick decides should the job be lauhcned at the tick
func (j *job) tick(t tick) bool {
	j.RLock()
	if _, ok := j.min[t.min]; !ok {
		j.RUnlock()
		return false
	}

	if _, ok := j.hour[t.hour]; !ok {
		j.RUnlock()
		return false
	}

	// cummulative day and weekDay, as it should be
	_, day := j.day[t.day]
	_, weekDay := j.weekDay[t.weekDay]
	if !day && !weekDay {
		j.RUnlock()
		return false
	}

	if _, ok := j.month[t.month]; !ok {
		j.RUnlock()
		return false
	}
	j.RUnlock()
	return true
}

func (j *job) run() {
	j.RLock()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[!] Cabbage Scheduler error, %v \n", r)
		}
	}()
	j.RUnlock()
	log.Printf("[*] scheduler publish task %s", j.taskName)
	j.publisher.PublishTask(j.taskName, j.fn())
}

// parseSchedule string and creates job struct with filled times to launch, or error if synthax is wrong
func parseSchedule(s string) (*job, error) {
	var err error
	j := &job{}
	// j.Lock()
	// defer j.Unlock()
	s = matchSpaces.ReplaceAllLiteralString(s, " ")
	parts := strings.Split(s, " ")
	if len(parts) != 5 {
		return j, errors.New("schedule string must have five components like * * * * *")
	}

	j.min, err = parsePart(parts[0], 0, 59)
	if err != nil {
		return j, err
	}

	j.hour, err = parsePart(parts[1], 0, 23)
	if err != nil {
		return j, err
	}

	j.day, err = parsePart(parts[2], 1, 31)
	if err != nil {
		return j, err
	}

	j.month, err = parsePart(parts[3], 1, 12)
	if err != nil {
		return j, err
	}

	j.weekDay, err = parsePart(parts[4], 0, 6)
	if err != nil {
		return j, err
	}

	//  day/weekDay combination
	switch {
	case len(j.day) < 31 && len(j.weekDay) == 7: // day set, but not weekDay, clear weekDay
		j.weekDay = make(map[int]struct{})
	case len(j.weekDay) < 7 && len(j.day) == 31: // weekDay set, but not day, clear day
		j.day = make(map[int]struct{})
	default:
		// both day and weekDay are * or both are set, use combined
		// i.e. don't do anything here
	}

	return j, nil
}

// parsePart parse individual schedule part from schedule string
func parsePart(s string, min, max int) (map[int]struct{}, error) {

	r := make(map[int]struct{})

	// wildcard pattern
	if s == "*" {
		for i := min; i <= max; i++ {
			r[i] = struct{}{}
		}
		return r, nil
	}

	// */2 1-59/5 pattern
	if matches := matchN.FindStringSubmatch(s); matches != nil {
		localMin := min
		localMax := max
		if matches[1] != "" && matches[1] != "*" {
			if rng := matchRange.FindStringSubmatch(matches[1]); rng != nil {
				localMin, _ = strconv.Atoi(rng[1])
				localMax, _ = strconv.Atoi(rng[2])
				if localMin < min || localMax > max {
					return nil, fmt.Errorf("out of range for %s in %s. %s must be in range %d-%d", rng[1], s, rng[1], min, max)
				}
			} else {
				return nil, fmt.Errorf("unable to parse %s part in %s", matches[1], s)
			}
		}
		n, _ := strconv.Atoi(matches[2])
		for i := localMin; i <= localMax; i += n {
			r[i] = struct{}{}
		}
		return r, nil
	}

	// 1,2,4  or 1,2,10-15,20,30-45 pattern
	parts := strings.Split(s, ",")
	for _, x := range parts {
		if rng := matchRange.FindStringSubmatch(x); rng != nil {
			localMin, _ := strconv.Atoi(rng[1])
			localMax, _ := strconv.Atoi(rng[2])
			if localMin < min || localMax > max {
				return nil, fmt.Errorf("iut of range for %s in %s. %s must be in range %d-%d", x, s, x, min, max)
			}
			for i := localMin; i <= localMax; i++ {
				r[i] = struct{}{}
			}
		} else if i, err := strconv.Atoi(x); err == nil {
			if i < min || i > max {
				return nil, fmt.Errorf("out of range for %d in %s. %d must be in range %d-%d", i, s, i, min, max)
			}
			r[i] = struct{}{}
		} else {
			return nil, fmt.Errorf("unable to parse %s part in %s", x, s)
		}
	}

	if len(r) == 0 {
		return nil, fmt.Errorf("unable to parse %s", s)
	}

	return r, nil
}

// getTick returns the tick struct from time
func getTick(t time.Time) tick {
	return tick{
		min:     t.Minute(),
		hour:    t.Hour(),
		day:     t.Day(),
		month:   int(t.Month()),
		weekDay: int(t.Weekday()),
	}
}
