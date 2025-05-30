package ulearning

type LoginResp struct {
	Authorization string `json:"AUTHORIZATION"`
	UserID        int    `json:"userId"`
	RoleID        int    `json:"roleId"`
}

type GetAllCoursesResp struct {
	CourseList []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		ClassID int    `json:"classId"`
	} `json:"courseList"`
}

type GetCourseActivitiesResp struct {
	OtherActivityDTOList []struct {
		RelationID   int    `json:"relationId"`
		RelationType int    `json:"relationType"`
		Title        string `json:"title"`
		PersonStatus int    `json:"personStatus"`
		Status       int    `json:"status"`
	} `json:"otherActivityDTOList"`
}

type GetActivityDetailResp struct {
	AbsenceNum     int    `json:"absenceNum"`
	NotAbsenceNum  int    `json:"notAbsenceNum"`
	Finish         string `json:"finish"`
	Location       string `json:"location"`
	AttendanceCode string `json:"attendanceCode"`
}

type SigninUser struct {
	UserID int `json:"userID"`
	Status int `json:"status"`
}

type SigninTeacherOPReq struct {
	AttendanceID int          `json:"attendanceID"`
	Users        []SigninUser `json:"users"`
}

type SigninTeacherOPResp struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

type SigninOperationReq struct {
	AttendanceID   int    `json:"attendanceID"`
	ClassID        int    `json:"classID"`
	UserID         int    `json:"userID"`
	Location       string `json:"location"`
	Address        string `json:"address"`
	EnterWay       int    `json:"enterWay"`
	AttendanceCode string `json:"attendanceCode"`
}

type SigninOperationResp struct {
	Msg       string `json:"msg"`
	NewStatus int    `json:"newStatus"`
	Status    int    `json:"status"`
}
