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
