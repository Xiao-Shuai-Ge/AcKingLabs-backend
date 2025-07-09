package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/manager"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域（生产环境需严格限制）
	},
}

func WebsocketAPI(c *gin.Context) {
	// 解析参数
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.WebsocketReq](c)
	if err != nil {
		return
	}
	// 验证token
	// 以空格分割，取出token
	list := strings.Split(req.Token, " ")
	if len(list) != 2 {
		zlog.CtxErrorf(ctx, "token格式错误")
		response.NewResponse(c).Error(response.TOKEN_FORMAT_ERROR)
		c.Abort()
		return
	}
	token := list[1]
	//解析token是否有效，并取出上一次的值
	data, err := jwtUtils.IdentifyToken(token)
	if err != nil {
		zlog.CtxErrorf(ctx, "token验证失败:%v", err)
		response.NewResponse(c).Error(response.TOKEN_IS_EXPIRED)
		//对应token无效，直接让他返回
		c.Abort()
		return
	}
	//判断其是否为atoken
	if data.Class != global.AUTH_ENUMS_ATOKEN {
		zlog.CtxErrorf(ctx, "token类型错误")
		response.NewResponse(c).Error(response.TOKEN_TYPE_ERROR)
		c.Abort()
		return
	}
	// 判断权限是否足够
	if data.Role < 1 {
		zlog.CtxErrorf(ctx, "权限不足")
		response.NewResponse(c).Error(response.PERMISSION_DENIED)
		c.Abort()
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 获取用户ID
	userIDStr := data.Userid
	// 转 int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		zlog.Errorf("%v 转换 int64 错误: %v", userIDStr, err)
		return
	}
	zlog.Infof("用户ID %d 连接 websocket", userID)

	// 注册客户端
	manager.WebsocketManager.Mutex.Lock()
	manager.WebsocketManager.Clients[conn] = userID
	manager.WebsocketManager.Users[userID] = conn
	manager.WebsocketManager.Mutex.Unlock()

	// 处理消息
	go func() {
		defer func() {
			manager.WebsocketManager.Mutex.Lock()
			delete(manager.WebsocketManager.Clients, conn)
			delete(manager.WebsocketManager.Users, userID)
			manager.WebsocketManager.Mutex.Unlock()

			zlog.Infof("用户ID %d 连接断开", userID)
			conn.Close()
		}()

		websocketLogic := logic.NewWebsocketLogic(conn)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 处理消息
			websocketLogic.HandleMessage(string(message))
		}
	}()

	//go Test()
}

func Test() {
	// 循环发送 10 次消息
	for i := 0; i < 10; i++ {
		content := types.ChatMessageResp{
			Type:      "chat",
			ID:        global.SnowflakeNode.Generate().Int64(),
			Content:   "hello world",
			UserID:    1908927295461003264,
			Timestamp: time.Now().UnixMilli(),
		}
		contentJson, _ := json.Marshal(&content)

		msg := manager.Message{
			Content: string(contentJson),
			ToType:  "all",
			To:      0,
		}
		manager.WebsocketManager.Broadcast <- msg
		time.Sleep(2 * time.Second)
	}
}
