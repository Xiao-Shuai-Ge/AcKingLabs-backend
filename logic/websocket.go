package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coze-dev/coze-go"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/manager"
	"tgwp/types"
	"time"
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
	// 分类解析消息
	switch data.Type {
	case "chat":
		l.handleTypeMessage(data.Content)
	default:
		zlog.Warnf("websocket 未知消息类型: %s", data.Type)
	}

	return
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
		UserID:    l.userID,
		Timestamp: time.Now().UnixMilli(),
	}
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
		go AiChat(data.Content)
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

func AiChat(content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	token := global.Config.Coze.Token
	botID := global.Config.Coze.BotID
	userID := "123456"

	authCli := coze.NewTokenAuth(token)

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL("https://api.coze.cn"), coze.WithHttpClient(&http.Client{
		Timeout: time.Minute * 2,
	}))

	// Step one, create chats
	req := &coze.CreateChatsReq{
		BotID:  botID,
		UserID: userID,
		Messages: []*coze.Message{
			coze.BuildUserQuestionText(content, nil),
		},
	}

	resp, err := cozeCli.Chat.Stream(ctx, req)
	if err != nil {
		fmt.Printf("Error starting chats: %v\n", err)
		return
	}

	Seq := 0                                      // 记录当前消息序号
	id := global.SnowflakeNode.Generate().Int64() // 生成唯一ID

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
		} else if event.Event == coze.ChatEventConversationChatCompleted {
			zlog.Debugf("本次使用token数: %d", event.Chat.Usage.TokenCount)
		} else {
			zlog.Debugf("未知事件: %s", event.Event)
		}
	}

	fmt.Printf("done, log:%s\n", resp.Response().LogID())
}
