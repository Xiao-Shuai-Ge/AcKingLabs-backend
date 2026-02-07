package logic

import (
	"context"
	"strconv"
	"tgwp/global"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"

	"tgwp/log/zlog"
)

type ReviewLogic struct {
}

func NewReviewLogic() *ReviewLogic {
	return &ReviewLogic{}
}

func (l *ReviewLogic) GetReviewList(ctx context.Context, req types.GetReviewListReq) (resp types.GetReviewListResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()

	offset := (req.Page - 1) * req.Count
	reviews, count, err := repo.NewReviewRepo(global.DB).GetReviewList(offset, req.Count)
	if err != nil {
		zlog.CtxErrorf(ctx, "GetReviewList failed: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	resp.Total = count
	var postIDs []int64
	for _, r := range reviews {
		postIDs = append(postIDs, r.PostID)
	}

	posts, err := repo.NewPostRepo(global.DB).GetPostsByIDs(postIDs)
	if err != nil {
		zlog.CtxErrorf(ctx, "GetPostsByIDs failed: %v", err)
	}

	for _, r := range reviews {
		detail := types.ReviewDetail{
			ID:         r.ID,
			PostID:     r.PostID,
			ReviewerID: r.ReviewerID,
			Status:     r.Status,
			Reason:     r.Reason,
			CreateTime: r.CreatedTime,
			UserID:     0,
		}
		if p, ok := posts[r.PostID]; ok {
			detail.PostTitle = p.Title
			detail.PostType = p.Type
			detail.UserID = p.UserID
		}
		resp.List = append(resp.List, detail)
	}
	return resp, nil
}

func (l *ReviewLogic) AuditReview(ctx context.Context, req types.AuditReviewReq) (resp types.AuditReviewResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()

	review, err := repo.NewReviewRepo(global.DB).GetReviewByID(req.ReviewID)
	if err != nil {
		zlog.CtxErrorf(ctx, "GetReviewByID failed: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	review.Status = req.Status
	// 解析 OperatorID
	opID, _ := strconv.ParseInt(req.OperatorID, 10, 64)
	review.ReviewerID = opID

	err = repo.NewReviewRepo(global.DB).UpdateReview(review)
	if err != nil {
		zlog.CtxErrorf(ctx, "UpdateReview failed: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	if req.Status == types.ReviewStatusPass {
		// 改为公开
		err = repo.NewPostRepo(global.DB).UpdatePostPrivate(review.PostID, false)
		if err != nil {
			zlog.CtxErrorf(ctx, "UpdatePostPrivate failed: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
	} else if req.Status == types.ReviewStatusReject {
		// 删除
		err = repo.NewPostRepo(global.DB).DeletePostByID(review.PostID)
		if err != nil {
			zlog.CtxErrorf(ctx, "DeletePostByID failed: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
	}

	return resp, nil
}
