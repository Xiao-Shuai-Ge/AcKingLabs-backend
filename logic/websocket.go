package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coze-dev/coze-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"strconv"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/manager"
	"tgwp/types"
	"time"
)

const (
	REDIS_AI_CONVERSATION_ID = "ai:conversation_id"
	REDIS_CHAT_MESSAGE_SET   = "chat:message:set"
)

type WebsocketLogic struct {
	conn   *websocket.Conn
	userID int64
}

func NewWebsocketLogic(conn *websocket.Conn) *WebsocketLogic {
	return &WebsocketLogic{
		conn:   conn,
		userID: manager.WebsocketManager.Clients[conn],
	}
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (l *WebsocketLogic) HandleMessage(message string) {
	var data Message
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		zlog.Warnf("websocket 接受消息格式错误: %s", err)
		return
	}
	zlog.Debugf("websocket 收到消息: %v", data)
	// 分类解析消息
	switch data.Type {
	case "chat":
		l.handleTypeMessage(data.Content)
	case "history":
		l.handleTypeGetHistory(data.Content)
	default:
		zlog.Warnf("websocket 未知消息类型: %s", data.Type)
	}

	return
}

func (l *WebsocketLogic) handleTypeGetHistory(content string) {
	// 处理消息内容
	var data types.GetHistoryReq
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		zlog.Warnf("websocket 接受消息格式错误: %s", err)
		return
	}
	// 从 redis 中获取历史消息
	res, err := global.Rdb.ZRevRangeByScore(context.Background(), REDIS_CHAT_MESSAGE_SET, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(data.Before-1, 10),
	}).Result()

	// 打包前 data.Count 条消息
	for i := 0; i < int(data.Count) && i < len(res); i++ {
		// 解析消息内容
		var resp types.ChatMessageResp
		err := json.Unmarshal([]byte(res[i]), &resp)
		if err != nil {
			zlog.Warnf("websocket 解析消息格式错误: %s", err)
			continue
		}
		resp.Type = "history"
		// 再次转换json
		respJson, err := json.Marshal(resp)
		if err != nil {
			zlog.Warnf("websocket 打包消息格式错误: %s", err)
			continue
		}
		// 单发消息
		msg := manager.Message{
			ToType:  "user",
			To:      l.userID,
			Content: string(respJson),
		}
		manager.WebsocketManager.Broadcast <- msg
	}
}

func (l *WebsocketLogic) handleTypeMessage(content string) {
	// 处理消息内容
	var data types.ChatMessageReq
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		zlog.Warnf("websocket 接受消息格式错误: %s", err)
		return
	}
	// 打包信息内容
	id := global.SnowflakeNode.Generate().Int64()
	resp := types.ChatMessageResp{
		Type:      "chat",
		ID:        id,
		Content:   data.Content,
		UserID:    strconv.FormatInt(l.userID, 10),
		Timestamp: time.Now().UnixMilli(),
	}
	SaveChatMessage(context.Background(), resp)
	respJson, err := json.Marshal(resp)
	if err != nil {
		zlog.Warnf("websocket 打包消息格式错误: %s", err)
		return
	}
	// 处理消息
	if data.Type == "all" {
		// 群发消息
		msg := manager.Message{
			ToType:  "all",
			Content: string(respJson),
		}
		manager.WebsocketManager.Broadcast <- msg
	} else if data.Type == "ai" {
		// 群发消息
		msg := manager.Message{
			ToType:  "all",
			Content: string(respJson),
		}
		manager.WebsocketManager.Broadcast <- msg
		// 触发AI回复
		go l.AiChat(data.Content)
	}
}

func SaveChatMessage(ctx context.Context, message types.ChatMessageResp) {
	// 转化 json 格式
	messageJson, err := json.Marshal(message)
	if err != nil {
		zlog.Errorf("websocket 打包消息格式错误: %s", err)
		return
	}
	//zlog.Debugf("websocket 保存消息: %s", messageJson)
	// 保存消息到 redis
	err = global.Rdb.ZAdd(ctx, REDIS_CHAT_MESSAGE_SET, &redis.Z{
		Score:  float64(message.Timestamp),
		Member: messageJson,
	}).Err()
	if err != nil {
		zlog.Errorf("websocket 保存消息到 redis 失败: %s", err)
		return
	}

	// 如果 redis 中的消息数量超过 50 条，清理最早的消息
	for global.Rdb.ZCount(ctx, REDIS_CHAT_MESSAGE_SET, "-inf", "+inf").Val() > 50 {
		err = global.Rdb.ZRemRangeByRank(ctx, REDIS_CHAT_MESSAGE_SET, 0, 0).Err()
		if err != nil {
			zlog.Errorf("websocket 清理消息到 redis 失败: %s", err)
			return
		}
	}
}

// 定义 API 请求和响应结构体
type AiMessage struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

type RequestBody struct {
	BotID              string      `json:"bot_id"`
	UserID             string      `json:"user_id"`
	Stream             bool        `json:"stream"`
	AdditionalMessages []AiMessage `json:"additional_messages"`
}

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (l *WebsocketLogic) AiChat(content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	token := global.Config.Coze.Token
	botID := global.Config.Coze.BotID
	userID := strconv.FormatInt(l.userID, 10)

	authCli := coze.NewTokenAuth(token)

	// 初始化 Coze API
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL("https://api.coze.cn"), coze.WithHttpClient(&http.Client{
		Timeout: time.Minute * 2,
	}))

	Seq := 0                                      // 记录当前消息序号
	id := global.SnowflakeNode.Generate().Int64() // 生成唯一ID
	conversationID := ""                          // 记录会话ID
	allContent := ""                              // 记录用户消息
	timestamp := time.Now().UnixMilli()           // 记录时间戳

	// 从 redis 中获取会话id
	value, err := global.Rdb.Get(ctx, REDIS_AI_CONVERSATION_ID).Result()
	if errors.Is(err, redis.Nil) {
		conversationID = ""
	} else if err != nil {
		return
	} else {
		zlog.Debugf("redis 获取会话id: %v", value)
		conversationID = value
	}
	zlog.Debugf("会话id: %v", conversationID)

	// 创建会话
	req := &coze.CreateChatsReq{
		ConversationID: conversationID,
		BotID:          botID,
		UserID:         userID,
		Messages: []*coze.Message{
			coze.BuildUserQuestionText(content, nil),
		},
	}

	resp, err := cozeCli.Chat.Stream(ctx, req)
	if err != nil {
		fmt.Printf("Error starting chats: %v\n", err)
		return
	}

	defer resp.Close()
	for {
		event, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			zlog.Debugf("开始流式传输")
			break
		}
		if err != nil {
			zlog.Errorf("流式传输错误: %v\n", err)
			break
		}
		if event.Event == coze.ChatEventConversationMessageDelta {
			if event.Message != nil && conversationID == "" {
				conversationID = event.Message.ConversationID
				zlog.Debugf("获取对话id: %v", conversationID)
				// 保存会话id到 redis
				err = global.Rdb.Set(ctx, REDIS_AI_CONVERSATION_ID, conversationID, time.Minute*5).Err()
				if err != nil {
					zlog.Errorf("保存会话id到 redis 失败: %v", err)
					return
				}
			}
			// 打包信息内容
			resp := types.AiMessageResp{
				Seq:       Seq,
				Type:      "ai",
				ID:        id,
				Content:   event.Message.Content,
				Timestamp: time.Now().UnixMilli(),
			}
			Seq++
			respJson, err := json.Marshal(resp)
			if err != nil {
				zlog.Errorf("websocket 打包消息格式错误: %s", err)
				return
			}
			// 群发传输信息
			msg := manager.Message{
				ToType:  "all",
				Content: string(respJson),
			}
			manager.WebsocketManager.Broadcast <- msg
			allContent += event.Message.Content
		} else if event.Event == coze.ChatEventConversationChatCompleted {
			zlog.Debugf("本次使用token数: %d", event.Chat.Usage.TokenCount)
		} else {
			zlog.Debugf("未知事件: %s", event.Event)
		}
	}

	// 保存聊天记录
	message := types.ChatMessageResp{
		Type:      "chat",
		ID:        id,
		Content:   allContent,
		UserID:    "ai",
		Timestamp: timestamp,
	}
	SaveChatMessage(context.Background(), message)
}
