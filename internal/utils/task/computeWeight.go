package task

import (
	"context"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/repo"
)

type ComputePostWeightPayload struct {
	PostID int64
}

func (d *Dispatcher) handleComputePostWeightJob(ctx context.Context, job Job) {
	if payload, ok := job.Payload.(ComputePostWeightPayload); ok {
		err := repo.NewPostRepo(global.DB).ComputePostWeightByID(payload.PostID)
		if err != nil {
			zlog.CtxErrorf(ctx, "Failed to compute post weight for ID %d: %v", payload.PostID, err)
		}
	} else {
		zlog.CtxErrorf(ctx, "Invalid payload for ComputePostWeight: %v", job.Payload)
	}
	zlog.CtxDebugf(ctx, "处理成功: %v", job.Payload)
}
