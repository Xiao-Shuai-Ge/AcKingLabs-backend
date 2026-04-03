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

func (d *Dispatcher) handleAIAuditJob(ctx context.Context, job Job) error {
	payload, ok := job.Payload.(AIAuditPayload)
	if !ok {
		return fmt.Errorf("invalid payload for AIAudit: %v", job.Payload)
	}
	workflowID := "7604072336694771762"
	params := map[string]interface{}{
		"content":     payload.Content,
		"sender_role": payload.SenderRole,
	}

	resultJSON, err := cozeUtils.RunWorkflow(ctx, workflowID, params)
	if err != nil {
		return fmt.Errorf("run AI audit workflow failed: %w", err)
	}

	zlog.CtxInfof(ctx, "AI audit workflow raw result: %s", resultJSON)

	var result AIAuditResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return fmt.Errorf("parse AI audit result failed: %w", err)
	}

	zlog.CtxInfof(ctx, "AI Audit Result - IsPass: %v, Reason: %s", result.IsPass, result.Reason)

	if !result.IsPass {
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
		if err := repo.NewReviewRepo(global.DB.WithContext(ctx)).CreateReview(review); err != nil {
			return fmt.Errorf("create review record failed: %w", err)
		}

		notifyContent := fmt.Sprintf("您的帖子因“%s”未通过审核，已隐藏并等待管理员二次审核。", result.Reason)
		url := fmt.Sprintf("/learn/%d", payload.PostID)
		notifyService.SendSystemMessageNotify(ctx, payload.UserID, notifyContent, url)

		if err := repo.NewPostRepo(global.DB).UpdatePostPrivate(payload.PostID, true); err != nil {
			return fmt.Errorf("hide post failed: %w", err)
		}
	}
	return nil
}
