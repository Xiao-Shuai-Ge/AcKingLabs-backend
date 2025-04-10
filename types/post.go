package types

type CreatePostReq struct {
	UserID string `json:"-"`

	Title   string `json:"title"`
	Content string `json:"content"`

	Type   string `json:"type"`
	Source string `json:"source"`

	IsPrivate bool `json:"is_private"`
}

type CreatePostResp struct {
	ID int64 `json:"id,string"`
}

type GetPostDetailReq struct {
	OperatorID   string `form:"-"`
	OperatorRole int    `form:"-"`

	ID string `form:"id"`
}

type GetPostDetailResp struct {
	ID        int64  `json:"id,string"`
	UserID    int64  `json:"user_id,string"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Likes     int    `json:"likes"`
	Comments  int    `json:"comments"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`

	IsAdminLike bool `json:"is_admin_like"`
	IsPrivate   bool `json:"is_private"`
	IsFeatured  bool `json:"is_featured"`
}
