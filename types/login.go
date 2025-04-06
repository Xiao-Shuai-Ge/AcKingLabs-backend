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

type LoginReq struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsRemember bool   `json:"is_remember"`
}

type LoginResp struct {
	Atoken string `json:"atoken"`
	Rtoken string `json:"rtoken"`
}

type RefreshTokenReq struct {
	Rtoken string `json:"rtoken"`
}

type RefreshTokenResp struct {
	Atoken string `json:"atoken"`
}

type TokenTestReq struct {
}

type TokenTestResp struct {
}
