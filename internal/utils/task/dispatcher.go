package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"tgwp/configs"
	"tgwp/log/zlog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	JOB_TYPE_COMPUTE_POST_WEIGHT = "ComputePostWeight"
	JOB_TYPE_AI_AUDIT            = "AIAudit"
	JOB_TYPE_AI_COMMENT          = "AIComment"
	JOB_TYPE_MENTION_NOTIFY      = "MentionNotify"

	retryHeaderKey = "x-retry-count"
	maxRetryCount  = 3
)

type Job struct {
	Type    string
	Payload interface{}
}

type Dispatcher struct {
	config     configs.RabbitMQConfig
	maxWorkers int
	conn       *amqp.Connection
	producer   *amqp.Channel
	consumer   *amqp.Channel
	deliveries <-chan amqp.Delivery
	wg         sync.WaitGroup
	mu         sync.RWMutex
	publishMu  sync.Mutex
	stopped    bool
}

type rabbitMQJobMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

var (
	GlobalDispatcher *Dispatcher
	once             sync.Once
)

func Init(cfg configs.RabbitMQConfig) {
	once.Do(func() {
		GlobalDispatcher = NewDispatcher(cfg)
		err := GlobalDispatcher.Run()
		if err != nil {
			panic(err)
		}
	})
}

func NewDispatcher(cfg configs.RabbitMQConfig) *Dispatcher {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	if cfg.PrefetchCount <= 0 {
		cfg.PrefetchCount = cfg.WorkerCount
	}
	if cfg.Exchange == "" {
		cfg.Exchange = "ackinglabs.task.exchange"
	}
	if cfg.Queue == "" {
		cfg.Queue = "ackinglabs.task.queue"
	}
	if cfg.RoutingKey == "" {
		cfg.RoutingKey = "ackinglabs.task"
	}
	if cfg.ConsumerTag == "" {
		cfg.ConsumerTag = "ackinglabs-task-consumer"
	}

	return &Dispatcher{
		config:     cfg,
		maxWorkers: cfg.WorkerCount,
	}
}

func (d *Dispatcher) Run() error {
	if !d.config.Enable {
		zlog.Warnf("RabbitMQ 未启用，任务调度器不会启动")
		return nil
	}

	if d.config.URL == "" {
		return fmt.Errorf("rabbitmq url 不能为空")
	}

	conn, err := amqp.Dial(d.config.URL)
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	producer, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("创建 RabbitMQ 生产通道失败: %w", err)
	}

	consumer, err := conn.Channel()
	if err != nil {
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("创建 RabbitMQ 消费通道失败: %w", err)
	}

	if err = producer.ExchangeDeclare(d.config.Exchange, "direct", true, false, false, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("声明 RabbitMQ exchange 失败: %w", err)
	}

	deadLetterExchange := d.deadLetterExchange()
	deadLetterQueue := d.deadLetterQueue()
	deadLetterRoutingKey := d.deadLetterRoutingKey()

	if err = producer.ExchangeDeclare(deadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("声明 RabbitMQ dead-letter exchange 失败: %w", err)
	}

	if _, err = producer.QueueDeclare(deadLetterQueue, true, false, false, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("声明 RabbitMQ dead-letter queue 失败: %w", err)
	}

	if err = producer.QueueBind(deadLetterQueue, deadLetterRoutingKey, deadLetterExchange, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("绑定 RabbitMQ dead-letter queue 失败: %w", err)
	}

	if _, err = producer.QueueDeclare(d.config.Queue, true, false, false, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("声明 RabbitMQ queue 失败: %w", err)
	}

	if err = producer.QueueBind(d.config.Queue, d.config.RoutingKey, d.config.Exchange, false, nil); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("绑定 RabbitMQ queue 失败: %w", err)
	}

	if err = consumer.Qos(d.config.PrefetchCount, 0, false); err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("设置 RabbitMQ qos 失败: %w", err)
	}

	deliveries, err := consumer.Consume(d.config.Queue, d.config.ConsumerTag, false, false, false, false, nil)
	if err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		_ = conn.Close()
		return fmt.Errorf("启动 RabbitMQ consumer 失败: %w", err)
	}

	d.conn = conn
	d.producer = producer
	d.consumer = consumer
	d.deliveries = deliveries

	for i := 0; i < d.maxWorkers; i++ {
		d.wg.Add(1)
		go d.worker()
	}

	zlog.Infof("RabbitMQ 任务调度器启动成功，worker: %d", d.maxWorkers)
	return nil
}

func (d *Dispatcher) Stop() {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.stopped = true
	d.mu.Unlock()

	if d.consumer != nil {
		if err := d.consumer.Close(); err != nil {
			zlog.Errorf("关闭 RabbitMQ 消费通道失败: %v", err)
		}
	}
	if d.producer != nil {
		if err := d.producer.Close(); err != nil {
			zlog.Errorf("关闭 RabbitMQ 生产通道失败: %v", err)
		}
	}
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			zlog.Errorf("关闭 RabbitMQ 连接失败: %v", err)
		}
	}

	d.wg.Wait()
}

func (d *Dispatcher) AddJob(job Job) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.stopped {
		zlog.CtxErrorf(context.Background(), "Dispatcher is stopped, job dropped: %v", job.Type)
		return
	}
	if !d.config.Enable {
		zlog.CtxErrorf(context.Background(), "RabbitMQ 未启用, job dropped: %v", job.Type)
		return
	}
	if d.producer == nil {
		zlog.CtxErrorf(context.Background(), "RabbitMQ producer 未初始化, job dropped: %v", job.Type)
		return
	}

	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		zlog.CtxErrorf(context.Background(), "任务 payload 序列化失败, job: %s, err: %v", job.Type, err)
		return
	}

	message := rabbitMQJobMessage{
		Type:    job.Type,
		Payload: payloadBytes,
	}
	body, err := json.Marshal(message)
	if err != nil {
		zlog.CtxErrorf(context.Background(), "任务消息序列化失败, job: %s, err: %v", job.Type, err)
		return
	}

	err = d.publishMessage(d.config.Exchange, d.config.RoutingKey, body, nil)
	if err != nil {
		zlog.CtxErrorf(context.Background(), "发布任务失败, job: %s, err: %v", job.Type, err)
	}
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for delivery := range d.deliveries {
		var message rabbitMQJobMessage
		err := json.Unmarshal(delivery.Body, &message)
		if err != nil {
			zlog.CtxErrorf(context.Background(), "反序列化任务消息失败: %v", err)
			_ = delivery.Ack(false)
			continue
		}

		job, err := d.parseJob(message)
		if err != nil {
			zlog.CtxErrorf(context.Background(), "解析任务消息失败: %v", err)
			_ = delivery.Ack(false)
			continue
		}

		retryCount := getRetryCount(delivery.Headers)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		err = d.handleJob(ctx, job)
		cancel()
		if err != nil {
			d.handleRetry(delivery, job, retryCount, err)
			continue
		}

		err = delivery.Ack(false)
		if err != nil {
			zlog.CtxErrorf(context.Background(), "任务消息确认失败: %v", err)
		}
	}
}

func (d *Dispatcher) parseJob(message rabbitMQJobMessage) (Job, error) {
	switch message.Type {
	case JOB_TYPE_COMPUTE_POST_WEIGHT:
		var payload ComputePostWeightPayload
		err := json.Unmarshal(message.Payload, &payload)
		if err != nil {
			return Job{}, fmt.Errorf("解析 ComputePostWeight payload 失败: %w", err)
		}
		return Job{Type: message.Type, Payload: payload}, nil
	case JOB_TYPE_AI_AUDIT:
		var payload AIAuditPayload
		err := json.Unmarshal(message.Payload, &payload)
		if err != nil {
			return Job{}, fmt.Errorf("解析 AIAudit payload 失败: %w", err)
		}
		return Job{Type: message.Type, Payload: payload}, nil
	case JOB_TYPE_AI_COMMENT:
		var payload AICommentPayload
		err := json.Unmarshal(message.Payload, &payload)
		if err != nil {
			return Job{}, fmt.Errorf("解析 AIComment payload 失败: %w", err)
		}
		return Job{Type: message.Type, Payload: payload}, nil
	case JOB_TYPE_MENTION_NOTIFY:
		var payload MentionNotifyPayload
		err := json.Unmarshal(message.Payload, &payload)
		if err != nil {
			return Job{}, fmt.Errorf("解析 MentionNotify payload 失败: %w", err)
		}
		return Job{Type: message.Type, Payload: payload}, nil
	default:
		return Job{}, fmt.Errorf("未知任务类型: %s", message.Type)
	}
}

func (d *Dispatcher) handleJob(ctx context.Context, job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker panic: %v", r)
		}
	}()

	switch job.Type {
	case JOB_TYPE_COMPUTE_POST_WEIGHT:
		return d.handleComputePostWeightJob(ctx, job)
	case JOB_TYPE_AI_AUDIT:
		return d.handleAIAuditJob(ctx, job)
	case JOB_TYPE_AI_COMMENT:
		return d.handleAICommentJob(ctx, job)
	case JOB_TYPE_MENTION_NOTIFY:
		return d.handleMentionNotifyJob(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}

	return nil
}

func (d *Dispatcher) handleRetry(delivery amqp.Delivery, job Job, retryCount int, jobErr error) {
	if retryCount < maxRetryCount {
		headers := cloneHeaders(delivery.Headers)
		headers[retryHeaderKey] = int32(retryCount + 1)
		err := d.publishMessage(d.config.Exchange, d.config.RoutingKey, delivery.Body, headers)
		if err != nil {
			zlog.CtxErrorf(context.Background(), "任务重试发布失败, job: %s, 当前重试次数: %d, err: %v", job.Type, retryCount, err)
			if nackErr := delivery.Nack(false, true); nackErr != nil {
				zlog.CtxErrorf(context.Background(), "任务消息拒绝失败, job: %s, err: %v", job.Type, nackErr)
			}
			return
		}
		zlog.CtxWarnf(context.Background(), "任务处理失败，准备重试, job: %s, 重试次数: %d/%d, err: %v", job.Type, retryCount+1, maxRetryCount, jobErr)
		if ackErr := delivery.Ack(false); ackErr != nil {
			zlog.CtxErrorf(context.Background(), "任务消息确认失败, job: %s, err: %v", job.Type, ackErr)
		}
		return
	}

	err := d.publishMessage(d.deadLetterExchange(), d.deadLetterRoutingKey(), delivery.Body, delivery.Headers)
	if err != nil {
		zlog.CtxErrorf(context.Background(), "任务投递死信队列失败, job: %s, err: %v", job.Type, err)
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			zlog.CtxErrorf(context.Background(), "任务消息拒绝失败, job: %s, err: %v", job.Type, nackErr)
		}
		return
	}

	zlog.CtxErrorf(context.Background(), "任务处理失败并进入死信队列, job: %s, 重试次数: %d, err: %v", job.Type, retryCount, jobErr)
	if ackErr := delivery.Ack(false); ackErr != nil {
		zlog.CtxErrorf(context.Background(), "任务消息确认失败, job: %s, err: %v", job.Type, ackErr)
	}
}

func (d *Dispatcher) publishMessage(exchange, routingKey string, body []byte, headers amqp.Table) error {
	d.publishMu.Lock()
	defer d.publishMu.Unlock()

	publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.producer.PublishWithContext(publishCtx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Timestamp:    time.Now(),
		Headers:      headers,
	})
}

func (d *Dispatcher) deadLetterExchange() string {
	return d.config.Exchange + ".dlx"
}

func (d *Dispatcher) deadLetterQueue() string {
	return d.config.Queue + ".dlq"
}

func (d *Dispatcher) deadLetterRoutingKey() string {
	return d.config.RoutingKey + ".dead"
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}

	value, ok := headers[retryHeaderKey]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	default:
		return 0
	}
}

func cloneHeaders(headers amqp.Table) amqp.Table {
	if len(headers) == 0 {
		return amqp.Table{}
	}

	cloned := make(amqp.Table, len(headers))
	for k, v := range headers {
		cloned[k] = v
	}
	return cloned
}
