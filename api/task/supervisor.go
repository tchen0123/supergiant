package task

import (
	"log"
	"time"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core"
)

const (
	interval = time.Second
)

type Supervisor struct {
	c    *core.Core
	pool []*worker
}

type worker struct {
}

type directive struct {
	task      *core.TaskResource
	performer Performable
}

func (_ *worker) work(ch <-chan *directive) {
	for {
		dir := <-ch
		task, performer := dir.task, dir.performer

		if err := task.Claim(); err != nil {
			// TODO the error here is presumed to be a CompareAndSwap error; if so,
			// we should just return. If it's another error, then this is not good.
			log.Println(err)
			continue
		}

		log.Printf("Starting %s task with ID %s", task.TypeName(), *task.ID)
		if err := performer.Perform(task.Data); err != nil {
			recordError(task, err)
			continue
		}

		log.Printf("Completed %s task with ID %s", task.TypeName(), *task.ID)
		task.Delete() // Task is successful, delete from Queue
	}
}

type Performable interface {
	Perform(data []byte) error
}

func NewSupervisor(c *core.Core, size int) *Supervisor {
	return &Supervisor{
		c:    c,
		pool: make([]*worker, size),
	}
}

func (s *Supervisor) Run() {

	// This starts all workers listening on the channel
	ch := make(chan *directive)
	for _, w := range s.pool {
		go w.work(ch)
	}

	ticker := time.NewTicker(interval)
	for _ = range ticker.C {

		tasks, err := s.c.Tasks().List()
		if err != nil {
			// panic(err)
			// TODO -- key does not exist, just keep going
			continue
		}

		// Find first queued task, or return.
		// Claim task and return if claim fails.
		var task *core.TaskResource
		for _, j := range tasks.Items {
			if j.IsQueued() {
				task = j
				break
			}
		}
		if task == nil {
			continue
		}

		var performer Performable
		switch task.Type {
		case common.TaskTypeDeleteApp:
			performer = DeleteApp{s.c}
		case common.TaskTypeDeleteComponent:
			performer = DeleteComponent{s.c}
		case common.TaskTypeDeployComponent:
			performer = DeployComponent{s.c}
		case common.TaskTypeStartInstance:
			performer = StartInstance{s.c}
		case common.TaskTypeStopInstance:
			performer = StopInstance{s.c}
		default:
			panic("Could not find task type")
		}

		ch <- &directive{task, performer}
	}
}

// Record error, and panic if that goes wrong
func recordError(task *core.TaskResource, err error) {
	if errRecordingErr := task.RecordError(err); errRecordingErr != nil {
		panic(errRecordingErr)
	}
}
