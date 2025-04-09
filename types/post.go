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
