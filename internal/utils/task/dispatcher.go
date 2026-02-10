package task

import (
	"context"
	"sync"
	"tgwp/log/zlog"
	"time"
)

const (
	JOB_TYPE_COMPUTE_POST_WEIGHT = "ComputePostWeight"
	JOB_TYPE_AI_AUDIT            = "AIAudit"
	JOB_TYPE_AI_COMMENT          = "AIComment"
)

type Job struct {
	Type    string
	Payload interface{}
}

type Dispatcher struct {
	jobQueue   chan Job
	maxWorkers int
	wg         sync.WaitGroup
	mu         sync.RWMutex
	stopped    bool
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
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

func (d *Dispatcher) Stop() {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.stopped = true
	close(d.jobQueue)
	d.mu.Unlock()

	d.wg.Wait()
}

func (d *Dispatcher) AddJob(job Job) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.stopped {
		zlog.CtxErrorf(context.Background(), "Dispatcher is stopped, job dropped: %v", job.Type)
		return
	}

	select {
	case d.jobQueue <- job:
	default:
		zlog.CtxErrorf(context.Background(), "Job queue is full, job dropped: %v", job.Type)
	}
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for job := range d.jobQueue {
		// 为每个任务创建一个 5 分钟超时的 context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		d.handleJob(ctx, job)
		cancel() // 任务执行完（或超时后 handleJob 返回）立即释放资源
	}
}

func (d *Dispatcher) handleJob(ctx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			zlog.CtxErrorf(ctx, "Worker panic: %v", r)
		}
	}()

	switch job.Type {
	case JOB_TYPE_COMPUTE_POST_WEIGHT:
		d.handleComputePostWeightJob(ctx, job)
	case JOB_TYPE_AI_AUDIT:
		d.handleAIAuditJob(ctx, job)
	case JOB_TYPE_AI_COMMENT:
		d.handleAICommentJob(ctx, job)
	default:
		zlog.CtxErrorf(ctx, "Unknown job type: %s", job.Type)
	}
}
