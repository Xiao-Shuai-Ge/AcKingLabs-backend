package task

import (
	"context"
	"fmt"
	"tgwp/global"
	"tgwp/repo"
)

type ComputePostWeightPayload struct {
	PostID int64
}

func (d *Dispatcher) handleComputePostWeightJob(ctx context.Context, job Job) error {
	payload, ok := job.Payload.(ComputePostWeightPayload)
	if !ok {
		return fmt.Errorf("invalid payload for ComputePostWeight: %v", job.Payload)
	}
	err := repo.NewPostRepo(global.DB.WithContext(ctx)).ComputePostWeightByID(payload.PostID)
	if err != nil {
		return fmt.Errorf("compute post weight failed for ID %d: %w", payload.PostID, err)
	}

	return nil
}
