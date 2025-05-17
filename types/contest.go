package types

type GetContestListReq struct {
	Type  string `form:"type"`
	Page  int    `form:"page"`
	Count int    `form:"count"`
}

type ContestInfo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Duration  int64  `json:"duration"`
	Platform  string `json:"platform"`
	Url       string `json:"url"`
}

type GetContestListResp struct {
	Contests  []ContestInfo `json:"contests"`
	Length    int           `json:"length"`
	PageTotal int64         `json:"page_total"`
}
