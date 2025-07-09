package types

type WebsocketReq struct {
	Token string `form:"token"`
}

type ChatMessageReq struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type ChatMessageResp struct {
	Type      string `json:"type"`
	ID        int64  `json:"id,string"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id,string"`
	Timestamp int64  `json:"timestamp"`
}
