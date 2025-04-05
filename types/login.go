package types

type SendCodeReq struct {
	Email string `json:"email"`
}

type SendCodeResp struct {
}

type RegisterReq struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type RegisterResp struct {
	Atoken string `json:"atoken"`
}

type TokenTestReq struct {
}

type TokenTestResp struct {
}
