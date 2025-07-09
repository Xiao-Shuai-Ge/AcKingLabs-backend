package manager

import (
	"github.com/gorilla/websocket"
	"sync"
	"tgwp/log/zlog"
)

var WebsocketManager *ClientManager

type Message struct {
	Content string `json:"content"` // 发送内容
	ToType  string `json:"to_type"` // 群发、私聊
	To      int64  `json:"to"`      // 发送对象 id
}

// 客户端连接管理
type ClientManager struct {
	Clients   map[*websocket.Conn]int64
	Users     map[int64]*websocket.Conn
	Broadcast chan Message
	Mutex     sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:   make(map[*websocket.Conn]int64),
		Users:     make(map[int64]*websocket.Conn),
		Broadcast: make(chan Message),
	}
}

func (cm *ClientManager) Start() {
	// 处理发送消息
	for {
		msg := <-cm.Broadcast
		zlog.Debugf("接受到消息：%v", msg)

		cm.Mutex.Lock()
		if msg.ToType == "all" {
			zlog.Debugf("群发消息: %v", msg.Content)
			for client := range cm.Clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(msg.Content))
				if err != nil {
					client.Close()
					delete(cm.Clients, client)
				}
			}
			cm.Mutex.Unlock()
		}
	}
}
