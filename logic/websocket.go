package logic

import (
	"encoding/json"
	"github.com/gorilla/websocket"
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
	}
}
