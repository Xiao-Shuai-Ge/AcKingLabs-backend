package types

type GetContestListReq struct {
	Type  string `form:"type"`
	Page  int    `form:"page"`
	Count int    `form:"count"`
}

type ContestInfo struct {
	ID          int64  `json:"id,string"`
	Title       string `json:"title"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Duration    int64  `json:"duration"`
	Platform    string `json:"platform"`
	Url         string `json:"url"`
	IsRecommend bool   `json:"is_recommend"`
}

type GetContestListResp struct {
	Contests  []ContestInfo `json:"contests"`
	Length    int           `json:"length"`
	PageTotal int64         `json:"page_total"`
}

type CreateContestReq struct {
	Title     string `json:"title"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Url       string `json:"url"`
}

type CreateContestResp struct {
	ID int64 `json:"id,string"`
}

type GetContestDetailReq struct {
	ContestID string `form:"contest_id"`
}

type GetContestDetailResp struct {
	Contest ContestInfo `json:"contest"`
}

type BookingContestReq struct {
	ContestID string `json:"contest_id"`
	UserID    string `json:"user_id"`
}

type BookingContestResp struct {
	IsBooking bool `json:"is_booking"`
}

type IsBookingContestReq struct {
	ContestID string `form:"contest_id"`
	UserID    string `form:"user_id"`
}

type IsBookingContestResp struct {
	IsBooking bool `json:"is_booking"`
}

type RecommendContestReq struct {
	ContestID   string `json:"contest_id"`
	IsRecommend bool   `json:"is_recommend"`
}

type RecommendContestResp struct {
	IsRecommend bool `json:"is_recommend"`
}
