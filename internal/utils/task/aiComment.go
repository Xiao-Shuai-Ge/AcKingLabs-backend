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

func (d *Dispatcher) handleAICommentJob(ctx context.Context, job Job) error {
	payload, ok := job.Payload.(AICommentPayload)
	if !ok {
		return fmt.Errorf("invalid payload for AIComment: %v", job.Payload)
	}
	workflowID := "7604364470764896319"
	params := map[string]interface{}{
		"content": payload.Content,
	}

	resultJSON, err := cozeUtils.RunWorkflow(ctx, workflowID, params)
	if err != nil {
		return fmt.Errorf("run AI comment workflow failed: %w", err)
	}

	zlog.CtxInfof(ctx, "AI comment workflow raw result: %s", resultJSON)

	var result AICommentResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return fmt.Errorf("parse AI comment result failed: %w", err)
	}

	zlog.CtxInfof(ctx, "AI Comment Result - Type: %v, Comment: %s", result.Type, result.Comment)

	if result.Type != 0 {
		comment := model.Comment{
			ID:       global.SnowflakeNode.Generate().Int64(),
			PostID:   payload.PostID,
			FatherID: 0,
			UserID:   1,
			Likes:    0,
			Content:  result.Comment,
			TimeModel: model.TimeModel{
				CreatedTime: time.Now().UnixMilli(),
				UpdatedTime: time.Now().UnixMilli(),
			},
		}
		if err := repo.NewPostRepo(global.DB.WithContext(ctx)).CreateComment(comment); err != nil {
			return fmt.Errorf("create auto comment failed: %w", err)
		}
		zlog.CtxInfof(ctx, "Successfully created auto comment for post %d", payload.PostID)

		post, err := repo.NewPostRepo(global.DB.WithContext(ctx)).GetPostDetail(payload.PostID)
		if err != nil {
			return fmt.Errorf("get post detail for notification failed: %w", err)
		}

		contentShort := result.Comment
		contentShort = strings.ReplaceAll(contentShort, "\n", " ")
		contentShort = utils.TruncateString(contentShort, 20)

		var url string
		if post.Type == "diary" {
			url = fmt.Sprintf("/diary/%d", post.ID)
		} else {
			url = fmt.Sprintf("/learn/%d", post.ID)
		}

		content := fmt.Sprintf("在你的帖子 《%s》 评论了: [%s]", post.Title, contentShort)
		notifyService.SendReplyNotify(ctx, post.UserID, 1, content, url)
	}
	return nil
}
