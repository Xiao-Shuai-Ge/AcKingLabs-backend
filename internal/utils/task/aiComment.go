package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"tgwp/global"
	"tgwp/internal/utils/notifyService"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/utils"
	"tgwp/utils/cozeUtils"
	"time"
)

type AICommentPayload struct {
	PostID  int64
	Content string
}

type AICommentResult struct {
	Type    int    `json:"type"`
	Comment string `json:"comment"`
}

func (d *Dispatcher) handleAICommentJob(ctx context.Context, job Job) {
	if payload, ok := job.Payload.(AICommentPayload); ok {
		workflowID := "7604364470764896319"
		params := map[string]interface{}{
			"content": payload.Content,
		}

		resultJSON, err := cozeUtils.RunWorkflow(ctx, workflowID, params)
		if err != nil {
			zlog.CtxErrorf(ctx, "Failed to run AI comment workflow: %v", err)
			return
		}

		zlog.CtxInfof(ctx, "AI comment workflow raw result: %s", resultJSON)

		var result AICommentResult
		// Coze workflow result might be a JSON string inside the data field
		if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
			zlog.CtxErrorf(ctx, "Failed to parse AI comment result: %v", err)
			return
		}

		zlog.CtxInfof(ctx, "AI Comment Result - Type: %v, Comment: %s", result.Type, result.Comment)

		// Type不为0时发布评论
		if result.Type != 0 {
			// 新建评论
			comment := model.Comment{
				ID:       global.SnowflakeNode.Generate().Int64(),
				PostID:   payload.PostID,
				FatherID: 0,
				UserID:   1, // 系统AI ID
				Likes:    0,
				Content:  result.Comment,
				TimeModel: model.TimeModel{
					CreatedTime: time.Now().UnixMilli(),
					UpdatedTime: time.Now().UnixMilli(),
				},
			}
			if err := repo.NewPostRepo(global.DB.WithContext(ctx)).CreateComment(comment); err != nil {
				zlog.CtxErrorf(ctx, "Failed to create auto comment: %v", err)
			} else {
				zlog.CtxInfof(ctx, "Successfully created auto comment for post %d", payload.PostID)

				// 获取帖子详情
				post, err := repo.NewPostRepo(global.DB.WithContext(ctx)).GetPostDetail(payload.PostID)
				if err != nil {
					zlog.CtxErrorf(ctx, "Failed to get post detail for notification: %v", err)
				} else {
					// 发送评论通知
					contentShort := result.Comment
					contentShort = strings.ReplaceAll(contentShort, "\n", " ")
					contentShort = utils.TruncateString(contentShort, 20)

					var url string
					if post.Type == "diary" {
						url = fmt.Sprintf("/diary/%d", post.ID)
					} else {
						url = fmt.Sprintf("/learn/%d", post.ID)
					}

					// 给帖子作者发送通知
					content := fmt.Sprintf("在你的帖子 《%s》 评论了: [%s]", post.Title, contentShort)
					notifyService.SendReplyNotify(ctx, post.UserID, 1, content, url)
				}
			}
		}

	} else {
		zlog.CtxErrorf(ctx, "Invalid payload for AIComment: %v", job.Payload)
	}
}
