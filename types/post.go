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

type EditPostReq struct {
	OperatorID   string `json:"-"`
	OperatorRole int    `json:"-"`

	PostID string `json:"post_id"`

	Title   string `json:"title"`
	Content string `json:"content"`

	Type   string `json:"type"`
	Source string `json:"source"`

	IsPrivate bool `json:"is_private"`
}

type EditPostResp struct {
}

type DeletePostReq struct {
	OperatorID   string `json:"-"`
	OperatorRole int    `json:"-"`

	PostID string `json:"post_id"`
}

type DeletePostResp struct {
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
	UserID   string `json:"-"`
	PostID   string `json:"post_id"`
	Content  string `json:"content"`
	FatherID string `json:"father_id"`
}

type CreateCommentResp struct {
	ID int64 `json:"id,string"`
}

type GetMoreCommentsReq struct {
	ID       string `form:"id"`
	IsChild  bool   `form:"is_child"`
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

type GetMorePostsReq struct {
	Type     string `form:"type"`
	Source   string `form:"source"`
	BeforeID string `form:"before_id"`
	By       string `form:"by"`
	Count    int    `form:"count"`
	UserID   string `form:"user_id"`
}

type PostInfo struct {
	ID           int64  `json:"id,string"`
	UserID       int64  `json:"user_id,string"`
	Title        string `json:"title"`
	ContentShort string `json:"content_short"`
	Type         string `json:"type"`
	Source       string `json:"source"`
	Likes        int    `json:"likes"`
	Comments     int    `json:"comments"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`

	IsAdminLike bool `json:"is_admin_like"`
	IsPrivate   bool `json:"is_private"`
	IsFeatured  bool `json:"is_featured"`

	Weight int64 `json:"weight"`
}

type GetMorePostsResp struct {
	Posts  []PostInfo `json:"posts"`
	Length int        `json:"length"`
}

type GetPagePostsReq struct {
	Type   string `form:"type"`
	Source string `form:"source"`
	Page   int    `form:"page"`
	By     string `form:"by"`
	Count  int    `form:"count"`
	UserID string `form:"user_id"`
}

type GetPagePostsResp struct {
	Posts     []PostInfo `json:"posts"`
	Length    int        `json:"length"`
	PageTotal int64      `json:"page_total"`
}

type SetPostFeatureReq struct {
	PostID string `json:"post_id"`
}

type SetPostFeatureResp struct {
}

type GetDiaryListReq struct {
	UserID string `form:"user_id"`
}

type DiaryInfo struct {
	PostID int64  `json:"post_id,string"`
	Source string `json:"source"`
}

type GetDiaryListResp struct {
	Posts  []DiaryInfo `json:"posts"`
	Length int         `json:"length"`
}

type SearchPostsReq struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Count   int    `form:"count"`
}

type SearchPostsResp struct {
	Posts     []PostInfo `json:"posts"`
	Length    int        `json:"length"`
	PageTotal int64      `json:"page_total"`
}
