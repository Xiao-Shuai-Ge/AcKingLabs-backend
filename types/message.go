package types

type GetMessageCountReq struct {
	UserID string `form:"-"`
}

type GetMessageCountResp struct {
	Count        int64 `json:"count"`
	SystemCount  int64 `json:"system_count"`
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
}

type GetMessageListReq struct {
	UserID string `form:"-"`
	Type   string `form:"type"`
}

type Message struct {
	ID        int64  `json:"id,string"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	SenderID  int64  `json:"sender_id,string"`
	Url       string `json:"url"`
	IsRead    bool   `json:"is_read"`
}

type GetMessageListResp struct {
	Messages []Message `json:"messages"`
	Length   int       `json:"length"`
}

type MarkReadMessageReq struct {
	UserID    string `json:"-"`
	MessageID string `json:"message_id"`
}

type MarkReadMessageResp struct {
}
