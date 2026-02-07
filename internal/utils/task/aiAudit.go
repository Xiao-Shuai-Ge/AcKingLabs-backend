package task

import (
	"context"
	"encoding/json"
	"fmt"
	"tgwp/global"
	"tgwp/internal/utils/notifyService"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/utils/cozeUtils"
	"time"
)

type AIAuditPayload struct {
	PostID     int64
	UserID     int64
	Title      string
	Content    string
	SenderRole int
}

type AIAuditResult struct {
	IsPass bool   `json:"is_pass"`
	Reason string `json:"reason"`
}

func (d *Dispatcher) handleAIAuditJob(ctx context.Context, job Job) {
	if payload, ok := job.Payload.(AIAuditPayload); ok {
		workflowID := "7604072336694771762"
		params := map[string]interface{}{
			"content":     payload.Content,
			"sender_role": payload.SenderRole,
		}

		resultJSON, err := cozeUtils.RunWorkflow(ctx, workflowID, params)
		if err != nil {
			zlog.CtxErrorf(ctx, "Failed to run AI audit workflow: %v", err)
			return
		}

		zlog.CtxInfof(ctx, "AI audit workflow raw result: %s", resultJSON)

		var result AIAuditResult
		// Coze workflow result might be a JSON string inside the data field
		if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
			zlog.CtxErrorf(ctx, "Failed to parse AI audit result: %v", err)
			return
		}

		zlog.CtxInfof(ctx, "AI Audit Result - IsPass: %v, Reason: %s", result.IsPass, result.Reason)

		// 如果审核不通过
		if !result.IsPass {
			// 1. 新建审核表记录
			review := model.Review{
				ID:         global.SnowflakeNode.Generate().Int64(),
				PostID:     payload.PostID,
				ReviewerID: 0,
				Status:     0,
				Reason:     result.Reason,
				TimeModel: model.TimeModel{
					CreatedTime: time.Now().UnixMilli(),
					UpdatedTime: time.Now().UnixMilli(),
				},
				PostTitle:   payload.Title,
				PostContent: payload.Content,
				UserID:      payload.UserID,
			}
			if err := repo.NewReviewRepo(global.DB).CreateReview(review); err != nil {
				zlog.CtxErrorf(ctx, "Failed to create review record: %v", err)
			}

			// 2. 发送通知给作者
			notifyContent := fmt.Sprintf("您的帖子因“%s”未通过审核，已隐藏并等待管理员二次审核。", result.Reason)
			// 这里 url 可以是帖子详情页或者特定的审核详情页，暂时跳转到帖子详情
			url := fmt.Sprintf("/learn/%d", payload.PostID)
			notifyService.SendSystemMessageNotify(ctx, payload.UserID, notifyContent, url)

			// 3. 将帖子改为隐藏
			if err := repo.NewPostRepo(global.DB).UpdatePostPrivate(payload.PostID, true); err != nil {
				zlog.CtxErrorf(ctx, "Failed to hide post: %v", err)
			}
		}

	} else {
		zlog.CtxErrorf(ctx, "Invalid payload for AIAudit: %v", job.Payload)
	}
}
