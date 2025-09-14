package types

// 投递简历请求
type SubmitResumeReq struct {
	RealName  string            `json:"real_name" binding:"required,min=2,max=20"`
	Grade     int               `json:"grade" binding:"required,min=0,max=99"`
	StudentNo string            `json:"student_no" binding:"required,min=1,max=30"`
	Email     string            `json:"email" binding:"required,email"`
	Extra     map[string]string `json:"extra" binding:"required"`
	Code      string            `json:"code" binding:"required,len=6"` // 邮箱验证码
}

// 投递简历响应
type SubmitResumeResp struct {
	ID int64 `json:"id,string"`
}

// 修改简历请求
type UpdateResumeReq struct {
	ID        string            `json:"id" binding:"required"`
	RealName  string            `json:"real_name" binding:"required,min=2,max=20"`
	Grade     int               `json:"grade" binding:"required,min=0,max=99"`
	StudentNo string            `json:"student_no" binding:"required,min=1,max=30"`
	Email     string            `json:"email" binding:"required,email"`
	Extra     map[string]string `json:"extra" binding:"required"`
	Code      string            `json:"code" binding:"required,len=6"` // 邮箱验证码
}

// 修改简历响应
type UpdateResumeResp struct {
}

// 查询简历详细信息请求
type GetResumeDetailReq struct {
	ID string `form:"id" binding:"required"`
}

// 查询简历详细信息响应
type GetResumeDetailResp struct {
	ID         int64             `json:"id,string"`
	RealName   string            `json:"real_name"`
	Grade      int               `json:"grade"`
	StudentNo  string            `json:"student_no"`
	Email      string            `json:"email"`
	Extra      map[string]string `json:"extra"`
	Code       string            `json:"code"`
	IsAccepted bool              `json:"is_accepted"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

// 获取简历列表请求（管理员功能）
type GetResumeListReq struct {
	Page  int `form:"page" binding:"required,min=1"`
	Count int `form:"count" binding:"required,min=1,max=100"`
}

// 简历列表项
type ResumeListItem struct {
	ID         int64  `json:"id,string"`
	RealName   string `json:"real_name"`
	Grade      int    `json:"grade"`
	StudentNo  string `json:"student_no"`
	Email      string `json:"email"`
	IsAccepted bool   `json:"is_accepted"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// 获取简历列表响应
type GetResumeListResp struct {
	Resumes   []ResumeListItem `json:"resumes"`
	Length    int              `json:"length"`
	PageTotal int64            `json:"page_total"`
	Total     int64            `json:"total"`
}

// 删除简历请求
type DeleteResumeReq struct {
	ID string `json:"id" binding:"required"`
}

// 删除简历响应
type DeleteResumeResp struct {
}

// 通过简历请求
type AcceptResumeReq struct {
	ID string `json:"id" binding:"required"`
}

// 通过简历响应
type AcceptResumeResp struct {
	Code string `json:"code"` // 生成的邀请码
}
