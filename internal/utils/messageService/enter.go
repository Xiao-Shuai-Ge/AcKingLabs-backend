package messageService

import (
	"context"
	"fmt"
	"sync"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/utils/cacheUtils"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeLike    MessageType = "like"    // 点赞消息
	MessageTypeComment MessageType = "comment" // 评论消息
	MessageTypeSystem  MessageType = "system"  // 系统消息
)

// MessageData 消息数据
type MessageData struct {
	UserID   int64       `json:"user_id"`   // 接收者ID
	SenderID int64       `json:"sender_id"` // 发送者ID
	Type     MessageType `json:"type"`      // 消息类型
	Content  string      `json:"content"`   // 消息内容
	URL      string      `json:"url"`       // 跳转链接
}

// MessageService 消息服务管理器
type MessageService struct {
	messageChan chan MessageData
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

var (
	messageService *MessageService
	once           sync.Once
)

// GetMessageService 获取消息服务单例
func GetMessageService() *MessageService {
	once.Do(func() {
		messageService = &MessageService{
			messageChan: make(chan MessageData, 100), // 缓冲100个消息
			workerCount: 3,                           // 3个工作协程
		}
	})
	return messageService
}

// Start 启动消息服务
func (ms *MessageService) Start() {
	ms.ctx, ms.cancel = context.WithCancel(context.Background())

	// 启动工作协程
	for i := 0; i < ms.workerCount; i++ {
		ms.wg.Add(1)
		go ms.worker(i)
	}

	zlog.Infof("消息服务已启动，工作协程数量: %d", ms.workerCount)
}

// Stop 停止消息服务
func (ms *MessageService) Stop() {
	if ms.cancel != nil {
		ms.cancel()
	}
	close(ms.messageChan)
	ms.wg.Wait()
	zlog.Infof("消息服务已停止")
}

// SendMessage 发送消息（异步）
func (ms *MessageService) SendMessage(data MessageData) {
	select {
	case ms.messageChan <- data:
		// 消息成功放入通道
	case <-time.After(5 * time.Second):
		// 超时，记录错误但不阻塞业务逻辑
		zlog.Errorf("发送消息超时: %+v", data)
	default:
		// 通道满了，记录错误但不阻塞业务逻辑
		zlog.Errorf("消息通道已满，丢弃消息: %+v", data)
	}
}

// worker 工作协程
func (ms *MessageService) worker(id int) {
	defer ms.wg.Done()

	zlog.Infof("消息处理工作协程 %d 已启动", id)

	for {
		select {
		case data := <-ms.messageChan:
			ms.processMessage(data)
		case <-ms.ctx.Done():
			zlog.Infof("消息处理工作协程 %d 已停止", id)
			return
		}
	}
}

// processMessage 处理单个消息
func (ms *MessageService) processMessage(data MessageData) {
	ctx := context.Background()

	// 生成消息ID
	messageID := global.SnowflakeNode.Generate().Int64()

	// 创建消息模型
	message := model.Message{
		ID:       messageID,
		UserID:   data.UserID,
		SenderID: data.SenderID,
		Type:     string(data.Type),
		Content:  data.Content,
		Url:      data.URL,
		IsRead:   false,
	}

	// 保存消息到数据库
	err := repo.NewMessageRepo(global.DB).SendMessage(message)
	if err != nil {
		zlog.CtxErrorf(ctx, "保存消息到数据库失败: %v, 消息数据: %+v", err, data)
		return
	}

	// 清理用户消息缓存
	cacheKey := fmt.Sprintf("cache:message_count:%d", data.UserID)
	err = cacheUtils.Remove(cacheKey)
	if err != nil {
		zlog.CtxErrorf(ctx, "清理消息缓存失败: %v, 用户ID: %d", err, data.UserID)
	}

	zlog.CtxInfof(ctx, "消息发送成功: 用户ID=%d, 类型=%s, 内容=%s",
		data.UserID, data.Type, data.Content)
}

// SendLikeMessage 发送点赞消息
func SendLikeMessage(userID, senderID int64, content, url string) {
	GetMessageService().SendMessage(MessageData{
		UserID:   userID,
		SenderID: senderID,
		Type:     MessageTypeLike,
		Content:  content,
		URL:      url,
	})
}

// SendCommentMessage 发送评论消息
func SendCommentMessage(userID, senderID int64, content, url string) {
	GetMessageService().SendMessage(MessageData{
		UserID:   userID,
		SenderID: senderID,
		Type:     MessageTypeComment,
		Content:  content,
		URL:      url,
	})
}

// SendSystemMessage 发送系统消息
func SendSystemMessage(userID int64, content, url string) {
	GetMessageService().SendMessage(MessageData{
		UserID:   userID,
		SenderID: 0, // 系统消息发送者ID为0
		Type:     MessageTypeSystem,
		Content:  content,
		URL:      url,
	})
}

// SendLikeMessageIfNotSelf 发送点赞消息（如果不是自己给自己点赞）
func SendLikeMessageIfNotSelf(userID, senderID int64, content, url string) {
	if userID != senderID {
		SendLikeMessage(userID, senderID, content, url)
	}
}

// SendCommentMessageIfNotSelf 发送评论消息（如果不是自己给自己评论）
func SendCommentMessageIfNotSelf(userID, senderID int64, content, url string) {
	if userID != senderID {
		SendCommentMessage(userID, senderID, content, url)
	}
}
