package task

import (
	"context"
	"sync"
	"tgwp/log/zlog"
)

const (
	JOB_TYPE_COMPUTE_POST_WEIGHT = "ComputePostWeight"
	JOB_TYPE_AI_AUDIT            = "AIAudit"
)

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
		d.handleComputePostWeightJob(ctx, job)
	case JOB_TYPE_AI_AUDIT:
		d.handleAIAuditJob(ctx, job)
	default:
		zlog.CtxErrorf(ctx, "Unknown job type: %s", job.Type)
	}
}
