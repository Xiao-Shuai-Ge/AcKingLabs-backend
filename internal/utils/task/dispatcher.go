package task

import (
	"context"
	"sync"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/repo"
)

const (
	JOB_TYPE_COMPUTE_POST_WEIGHT = "ComputePostWeight"
)

type ComputePostWeightPayload struct {
	PostID int64
}

type Job struct {
	Type    string
	Payload interface{}
}

type Dispatcher struct {
	jobQueue   chan Job
	maxWorkers int
	wg         sync.WaitGroup
	quit       chan bool
}

var (
	GlobalDispatcher *Dispatcher
	once             sync.Once
)

func Init(maxWorkers, jobQueueSize int) {
	once.Do(func() {
		GlobalDispatcher = NewDispatcher(maxWorkers, jobQueueSize)
		GlobalDispatcher.Run()
	})
}

func NewDispatcher(maxWorkers int, jobQueueSize int) *Dispatcher {
	return &Dispatcher{
		jobQueue:   make(chan Job, jobQueueSize),
		maxWorkers: maxWorkers,
		quit:       make(chan bool),
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

func (d *Dispatcher) Stop() {
	close(d.quit)
	d.wg.Wait()
}

func (d *Dispatcher) AddJob(job Job) {
	d.jobQueue <- job
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for {
		select {
		case job := <-d.jobQueue:
			d.handleJob(job)
		case <-d.quit:
			return
		}
	}
}

func (d *Dispatcher) handleJob(job Job) {
	defer func() {
		if r := recover(); r != nil {
			ctx := context.Background()
			zlog.CtxErrorf(ctx, "Worker panic: %v", r)
		}
	}()

	ctx := context.Background()
	switch job.Type {
	case JOB_TYPE_COMPUTE_POST_WEIGHT:
		if payload, ok := job.Payload.(ComputePostWeightPayload); ok {
			err := repo.NewPostRepo(global.DB).ComputePostWeightByID(payload.PostID)
			if err != nil {
				zlog.CtxErrorf(ctx, "Failed to compute post weight for ID %d: %v", payload.PostID, err)
			}
		} else {
			zlog.CtxErrorf(ctx, "Invalid payload for ComputePostWeight: %v", job.Payload)
		}
		zlog.CtxDebugf(ctx, "处理成功: %v", job.Payload)
	default:
		zlog.CtxErrorf(ctx, "Unknown job type: %s", job.Type)
	}
}
