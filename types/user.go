package types

type GetUserInfoReq struct {
	ID string `form:"id"`
}

type GetUserInfoResp struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Xp       int    `json:"xp"`
	Role     int    `json:"role"`
}

type GetUserProfileReq struct {
	ID string `form:"id"`
}

type GetUserProfileResp struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Xp       int    `json:"xp"`

	Grade     int    `json:"grade"`
	StudentNo string `json:"student_no"`
	RealName  string `json:"real_name"`

	CodeforcesID     string `json:"codeforces_id"`
	CodeforcesRating int    `json:"codeforces_rating"`

	Role int `json:"role"`
}

type SetUserProfileReq struct {
	OperatorID   string `json:"-"`
	OperatorRole int    `json:"-"`

	ID string `json:"id"`

	Username string `json:"username"`
	Avatar   string `json:"avatar"`

	Grade     int    `json:"grade"`
	StudentNo string `json:"student_no"`
	RealName  string `json:"real_name"`

	CodeforcesID string `json:"codeforces_id"`
}

type SetUserProfileResp struct {
}

type SetUserRoleReq struct {
	OperatorRole int `json:"-"`

	ID   string `json:"id"`
	Role int    `json:"role"`
}

type SetUserRoleResp struct {
}

type GetRankingsReq struct {
	Page  int `form:"page"`
	Count int `form:"count"`
}

type Ranking struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Xp       int    `json:"xp"`
	Role     int    `json:"role"`
}

type GetRankingsResp struct {
	Rankings  []Ranking `json:"rankings"`
	Length    int       `json:"length"`
	PageTotal int64     `json:"page_total"`
}

// 删除用户请求
type DeleteUserReq struct {
	ID string `json:"id"`
}

// 删除用户响应
type DeleteUserResp struct {
}

// 获取用户列表请求（按ID排序分页）
type GetUserListReq struct {
	Page  int `form:"page" binding:"required,min=1"`
	Count int `form:"count" binding:"required,min=1,max=100"`
}

// 用户列表项
type UserListItem struct {
	ID        int64  `json:"id,string"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Xp        int    `json:"xp"`
	Grade     int    `json:"grade"`
	RealName  string `json:"real_name"`
	Role      int    `json:"role"`
	CreatedAt string `json:"created_at"`
}

// 获取用户列表响应
type GetUserListResp struct {
	Users     []UserListItem `json:"users"`
	Length    int            `json:"length"`
	PageTotal int64          `json:"page_total"`
	Total     int64          `json:"total"`
}
