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
	UserID    string `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

type AiMessageResp struct {
	Seq       int    `json:"seq"`
	Type      string `json:"type"`
	ID        int64  `json:"id,string"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type GetHistoryReq struct {
	Before int64 `json:"before"`
	Count  int64 `json:"count"`
}
