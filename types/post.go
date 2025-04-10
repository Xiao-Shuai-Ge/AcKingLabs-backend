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

type LikePostReq struct {
	OperatorID   string `form:"-"`
	OperatorRole int    `form:"-"`

	PostID string `json:"post_id"`
}

type LikePostResp struct {
	IsLike bool `json:"is_like"`
}

type GetLikePostReq struct {
	OperatorID string `form:"-"`
	PostID     string `form:"post_id"`
}

type GetLikePostResp struct {
	IsLike bool `json:"is_like"`
}

type CreateCommentReq struct {
	UserID  string `json:"-"`
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}

type CreateCommentResp struct {
	ID int64 `json:"id,string"`
}

type GetMoreCommentsReq struct {
	PostID   string `form:"post_id"`
	BeforeID string `form:"before_id"`
	Count    int    `form:"count"`
}

type Comment struct {
	ID          int64  `json:"id,string"`
	UserID      int64  `json:"user_id,string"`
	Content     string `json:"content"`
	Likes       int    `json:"likes"`
	CreatedAt   int64  `json:"created_at"`
	IsAdminLike bool   `json:"is_admin_like"`
}

type GetMoreCommentsResp struct {
	Comments []Comment `json:"comments"`
	Length   int       `json:"length"`
}

type LikeCommentReq struct {
	OperatorID   string `form:"-"`
	OperatorRole int    `form:"-"`

	CommentID string `json:"comment_id"`
}

type LikeCommentResp struct {
	IsLike bool `json:"is_like"`
}

type GetLikeCommentReq struct {
	OperatorID string `form:"-"`
	CommentID  string `form:"comment_id"`
}

type GetLikeCommentResp struct {
	IsLike bool `json:"is_like"`
}
