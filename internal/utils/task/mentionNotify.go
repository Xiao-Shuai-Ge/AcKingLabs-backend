package task

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"tgwp/global"
	"tgwp/internal/utils/notifyService"
	"tgwp/log/zlog"
	"tgwp/repo"
)

type MentionNotifyPayload struct {
	PostID     int64  // 帖子ID
	SenderID   int64  // 发送者ID
	Content    string // 内容
	SourceType string // 来源类型：post, comment
}

func (d *Dispatcher) handleMentionNotifyJob(ctx context.Context, job Job) {
	payload, ok := job.Payload.(MentionNotifyPayload)
	if !ok {
		zlog.CtxErrorf(ctx, "Invalid payload for MentionNotify: %v", job.Payload)
		return
	}

	zlog.CtxInfof(ctx, "Start processing MentionNotify job for %s %d", payload.SourceType, payload.PostID)

	// 1. 正则匹配所有 @的用户
	// 格式：[@{用户名}](/profile/{用户id})
	re := regexp.MustCompile(`\[@(.*?)\]\(/profile/(\d+)\)`)
	matches := re.FindAllStringSubmatch(payload.Content, -1)

	if len(matches) == 0 {
		return
	}

	// 2. 去重并收集用户ID
	userIDs := make(map[int64]string) // map[UserID]UserName
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		username := match[1]
		userIDStr := match[2]
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			continue
		}
		// 不通知自己
		if userID == payload.SenderID {
			continue
		}
		userIDs[userID] = username
	}

	if len(userIDs) == 0 {
		return
	}

	// 3. 获取帖子信息用于生成链接和标题
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(payload.PostID)
	if err != nil {
		zlog.CtxErrorf(ctx, "Failed to get post detail: %v", err)
		return
	}

	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}

	// 4. 发送通知
	for userID, username := range userIDs {
		// 检查用户是否存在
		_, err := repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
		if err != nil {
			zlog.CtxWarnf(ctx, "User %d not found, skip mention notify", userID)
			continue
		}

		var content string
		if payload.SourceType == "post" {
			content = fmt.Sprintf("在帖子《%s》中提到了你", post.Title)
		} else {
			// 帖子标题作为上下文
			content = fmt.Sprintf("在帖子《%s》的评论中提到了你", post.Title)
		}

		notifyService.SendMentionNotify(ctx, userID, payload.SenderID, content, url)
		zlog.CtxInfof(ctx, "Sent mention notify to user %s(%d)", username, userID)
	}
}
