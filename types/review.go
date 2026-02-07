package types

type GetReviewListReq struct {
	Page         int    `json:"page" form:"page" binding:"required,min=1"`
	Count        int    `json:"count" form:"count" binding:"required,min=1,max=100"`
	OperatorID   string `json:"operator_id" form:"operator_id"`
	OperatorRole int    `json:"operator_role" form:"operator_role"`
}

type GetReviewListResp struct {
	List  []ReviewDetail `json:"list"`
	Total int64          `json:"total"`
}

type ReviewDetail struct {
	ID         int64  `json:"id,string"`
	PostID     int64  `json:"post_id,string"`
	ReviewerID int64  `json:"reviewer_id,string"`
	Status     int    `json:"status"`
	Reason     string `json:"reason"`
	CreateTime int64  `json:"create_time"`
	PostTitle  string `json:"post_title"`
	PostType   string `json:"post_type"`
	UserID     int64  `json:"user_id,string"`
}

const (
	ReviewStatusPass   = 1
	ReviewStatusReject = 2
)

type AuditReviewReq struct {
	ReviewID     int64  `json:"review_id,string" binding:"required"`
	Status       int    `json:"status" binding:"required,oneof=1 2"` // 1: Pass, 2: Reject
	OperatorID   string `json:"operator_id" form:"operator_id"`
	OperatorRole int    `json:"operator_role" form:"operator_role"`
}

type AuditReviewResp struct {
}
