package types

type TemplateReq struct {
}

type TemplateResp struct {
}

type SigninListReq struct {
	ID string `form:"-"`
}

type Active struct {
	UserID        int    `json:"user_id"`
	ClassName     string `json:"class_name"`
	ActiveName    string `json:"active_name"`
	ClassID       int    `json:"class_id"`
	RelationID    int    `json:"relation_id"`
	AbsenceNum    int    `json:"absence_num"`
	NotAbsenceNum int    `json:"not_absence_num"`
	IsFinished    bool   `json:"is_finished"`
}

type Course struct {
	CourseID   int    `json:"course_id"`
	CourseName string `json:"course_name"`
	ClassID    int    `json:"class_id"`
}

type SigninListResp struct {
	Actives []Active `json:"actives"`
	Courses []Course `json:"courses"`
	Length  int      `json:"length"`
	Level   int      `json:"level"`
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

type GetAutoListReq struct {
	UserID string `json:"-"`
}

type AutoCourse struct {
	CourseID int `json:"course_id"`
	Percent  int `json:"percent"`
}

type GetAutoListResp struct {
	CourseList []AutoCourse `json:"course_list"`
}

type AutoSettingReq struct {
	UserID     string `json:"-"`
	IsAuto     bool   `json:"is_auto"`
	ClassID    int    `json:"class_id"`
	CourseName string `json:"course_name"`
	CourseID   int    `json:"course_id"`
	Percent    int    `json:"percent"`
}

type AutoSettingResp struct {
	IsAuto bool `json:"is_auto"`
}
