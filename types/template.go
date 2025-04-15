package types

type TemplateReq struct {
}

type TemplateResp struct {
}

type SigninListReq struct {
	ID string `form:"-"`
}

type Active struct {
	UserID     int    `json:"user_id"`
	ClassName  string `json:"class_name"`
	ActiveName string `json:"active_name"`
	ClassID    int    `json:"class_id"`
	RelationID int    `json:"relation_id"`
}

type SigninListResp struct {
	Actives []Active `json:"actives"`
	Length  int      `json:"length"`
}

type SigninReq struct {
	ID         string `json:"-"`
	UserID     int    `json:"user_id"`
	ClassID    int    `json:"class_id"`
	RelationID int    `json:"relation_id"`
}

type SigninResp struct {
	Message string `json:"message"`
}

type SigninTeacherReq struct {
	ID         string `json:"-"`
	UserID     int    `json:"user_id"`
	RelationID int    `json:"relation_id"`
}

type SigninTeacherResp struct {
	Message string `json:"message"`
}
