package types

type TemplateReq struct {
	Body string `form:"body"`
}

type TemplateResp struct {
	Body string `json:"body"`
}
