package cozeUtils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tgwp/global"
	"tgwp/log/zlog"
	"time"

	"github.com/coze-dev/coze-go"
)

// StreamCallback defines the function signature for handling stream events
// content: the delta content received
// conversationID: the conversation ID
// isDelta: true if this is a message delta
// usage: token usage info (only present when completed)
type StreamCallback func(content string, conversationID string, isDelta bool, usage *coze.ChatUsage) error

// ChatStream initiates a streaming chat with Coze
func ChatStream(ctx context.Context, userID string, conversationID string, content string, callback StreamCallback) (string, error) {
	token := global.Config.Coze.Token
	botID := global.Config.Coze.BotID

	authCli := coze.NewTokenAuth(token)

	// Initialize Coze API
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL("https://api.coze.cn"), coze.WithHttpClient(&http.Client{
		Timeout: time.Minute * 2,
	}))

	// Create chat request
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
		return conversationID, err
	}

	defer resp.Close()

	finalConversationID := conversationID

	for {
		event, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			zlog.Debugf("开始流式传输结束")
			break
		}
		if err != nil {
			zlog.Errorf("流式传输错误: %v\n", err)
			return finalConversationID, err
		}

		if event.Event == coze.ChatEventConversationMessageDelta {
			if event.Message != nil {
				if finalConversationID == "" {
					finalConversationID = event.Message.ConversationID
				}
				if err := callback(event.Message.Content, finalConversationID, true, nil); err != nil {
					return finalConversationID, err
				}
			}
		} else if event.Event == coze.ChatEventConversationChatCompleted {
			if event.Chat != nil && event.Chat.Usage != nil {
				if err := callback("", finalConversationID, false, event.Chat.Usage); err != nil {
					return finalConversationID, err
				}
			}
		} else {
			zlog.Debugf("未知事件: %s", event.Event)
		}
	}

	return finalConversationID, nil
}
