package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
)

type TemplateLogic struct {
}

func NewTemplateLogic() *TemplateLogic {
	return &TemplateLogic{}
}

// 这个包内的常量
const (
	REDIS_SNOW_ID = "island:test.code:string"
)

func (l *TemplateLogic) Way(ctx context.Context, req types.TemplateReq) (resp types.TemplateResp, err error) {
	defer utils.RecordTime(time.Now())()

	return
}

func (l *TemplateLogic) SigninList(ctx context.Context, req types.SigninListReq) (resp types.SigninListResp, err error) {
	defer utils.RecordTime(time.Now())()
	// id 转 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库拿去用户账号密码
	userInfo, err := repo.NewTemplateRepo(global.DB).GetUserInfo(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户 %d 不在白名单中，权限不足", userID)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 模拟登录
	token, UserID, err := Login(userInfo.UserName, userInfo.Password)
	zlog.CtxDebugf(ctx, "模拟登录成功: token=%s, userID=%d", token, userID)

	// 获取课程列表
	var getAllCoursesResp GetAllCoursesResp
	getAllCoursesResp, err = GetAllCourses(token)
	if err != nil {
		zlog.Errorf("获取课程列表失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	//zlog.CtxDebugf(ctx, "获取课程列表成功: %v", getAllCoursesResp)

	for _, course := range getAllCoursesResp.CourseList {
		// zlog.CtxDebugf(ctx, "课程名称: %s", course.Name)
		// 获取课程活动
		var getCourseActivitiesResp GetCourseActivitiesResp
		getCourseActivitiesResp, err = GetCourseActivities(token, course.ID)
		if err != nil {
			zlog.Errorf("获取课程活动失败: %v", err)
			return resp, response.ErrResp(err, response.INTERNAL_ERROR)
		}
		//zlog.CtxDebugf(ctx, "获取课程活动成功: %v", getCourseActivitiesResp)
		for _, activity := range getCourseActivitiesResp.OtherActivityDTOList {
			//zlog.CtxDebugf(ctx, "课程活动名称: %s", activity.Title)
			if activity.RelationType == 1 && activity.PersonStatus != 1 {
				zlog.CtxDebugf(ctx, "课程活动名称: %s", activity.Title)
				Active := types.Active{
					UserID:     UserID,
					ClassName:  course.Name,
					ActiveName: activity.Title,
					ClassID:    course.ClassID,
					RelationID: activity.RelationID,
				}
				resp.Actives = append(resp.Actives, Active)
			}
		}
	}
	resp.Length = len(resp.Actives)
	return
}

func (l *TemplateLogic) Signin(ctx context.Context, req types.SigninReq) (resp types.SigninResp, err error) {
	defer utils.RecordTime(time.Now())()
	// id 转 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库拿去用户账号密码
	userInfo, err := repo.NewTemplateRepo(global.DB).GetUserInfo(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户 %d 不在白名单中，权限不足", userID)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 模拟登录
	token, UserID, err := Login(userInfo.UserName, userInfo.Password)
	zlog.CtxDebugf(ctx, "模拟登录成功: token=%s, userID=%d", token, userID)

	// 获取活动信息
	var signinDetailResp SigninDetailResp
	signinDetailResp, err = GetSigninDetail(token, req.RelationID)
	if err != nil {
		zlog.Errorf("获取活动信息失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	zlog.CtxDebugf(ctx, "获取活动信息成功: %v", signinDetailResp)
	// 嘿嘿嘿
	var signinOperationResp SigninOperationResp
	signinOperationResp, err = SigninOperation(SigninOperationData{
		token:          token,
		attendanceType: signinDetailResp.Type,
		relatedID:      req.RelationID,
		classID:        req.ClassID,
		userID:         UserID,
		location:       signinDetailResp.Location,
		attendanceCode: signinDetailResp.AttendanceCode,
		force:          false,
	})
	if err != nil {
		zlog.Errorf("签到失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	msg := fmt.Sprintf("[ %s ] 签到效果: %v", signinDetailResp.Title, signinOperationResp.Msg)
	if signinOperationResp.NewStatus != 1 {
		signinOperationResp, err = SigninOperation(SigninOperationData{
			token:          token,
			attendanceType: signinDetailResp.Type,
			relatedID:      req.RelationID,
			classID:        req.ClassID,
			userID:         UserID,
			location:       signinDetailResp.Location,
			attendanceCode: signinDetailResp.AttendanceCode,
			force:          true,
		})
		msg += fmt.Sprintf(" [ %s ] 再次签到效果: %v", signinDetailResp.Title, signinOperationResp.Msg)
		if signinOperationResp.NewStatus != 1 {
			zlog.CtxErrorf(ctx, "签到失败: %v", err)
			return resp, response.ErrResp(err, response.INTERNAL_ERROR)
		}
	}
	zlog.CtxInfof(ctx, "签到成功: %v", msg)

	resp.Message = msg
	return
}

type SigninOperationData struct {
	attendanceType int
	token          string
	relatedID      int
	relatedType    int
	classID        int
	userID         int
	location       string
	attendanceCode string
	force          bool
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

func SigninOperation(data SigninOperationData) (resp SigninOperationResp, err error) {
	zlog.CtxDebugf(context.Background(), "开始签到: %v", data)
	var reqData = SigninOperationReq{
		AttendanceID:   data.relatedID,
		ClassID:        data.classID,
		UserID:         data.userID,
		Location:       data.location,
		Address:        "",
		EnterWay:       1,
		AttendanceCode: data.attendanceCode,
	}
	zlog.CtxDebugf(context.Background(), "开始签到: %v", reqData)
	if data.attendanceType == 0 {
		reqData.AttendanceCode = ""
	} else if data.attendanceType == 1 || data.attendanceType == 2 {
		reqData.Location = ""
	}
	if data.force {
		reqData.Location = "113,22"
		reqData.AttendanceCode = ""
	}

	jsonData, _ := json.Marshal(reqData)

	zlog.CtxDebugf(context.Background(), "请求签到数据: %s", string(jsonData))

	req, _ := http.NewRequest("POST", fmt.Sprintf(global.Config.Ulearning.SigninOperation), bytes.NewBuffer(jsonData))
	req.Header.Add("Authorization", data.token)
	req.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Errorf("请求签到失败: %v", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	zlog.CtxDebugf(context.Background(), "签到结果: %s", string(body))
	json.Unmarshal([]byte(body), &resp)

	return
}

type SigninDetailResp struct {
	Location       string `json:"location"`
	AttendanceCode string `json:"attendanceCode"`
	Type           int    `json:"type"`
	Title          string `json:"title"`
}

func GetSigninDetail(token string, relatedID int) (resp SigninDetailResp, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(global.Config.Ulearning.GetSigninDetail, relatedID), nil)
	req.Header.Add("Authorization", token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Errorf("请求签到详情失败: %v", err)
		return
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	zlog.CtxDebugf(context.Background(), "签到详情: %s", string(body))
	json.Unmarshal([]byte(body), &resp)

	return
}

type GetCourseActivitiesResp struct {
	OtherActivityDTOList []struct {
		RelationID   int    `json:"relationId"`
		RelationType int    `json:"relationType"`
		Title        string `json:"title"`
		PersonStatus int    `json:"personStatus"`
	} `json:"otherActivityDTOList"`
}

func GetCourseActivities(token string, courseID int) (resp GetCourseActivitiesResp, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(global.Config.Ulearning.GetCourseActivities, courseID), nil)
	req.Header.Add("Authorization", token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Errorf("请求课程活动失败: %v", err)
		return
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	json.Unmarshal([]byte(body), &resp)

	return
}

type GetAllCoursesResp struct {
	CourseList []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		ClassID int    `json:"classId"`
	} `json:"courseList"`
}

// GetAllCourses 获取课程列表
func GetAllCourses(token string) (resp GetAllCoursesResp, err error) {
	req, _ := http.NewRequest("GET", global.Config.Ulearning.GetAllCourses, nil)
	req.Header.Add("Authorization", token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Errorf("请求课程列表失败: %v", err)
		return
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	json.Unmarshal([]byte(body), &resp)

	//zlog.Debugf("课程列表: %v", resp)
	return
}

type LoginResponse struct {
	Authorization string `json:"AUTHORIZATION"`
	UserID        int    `json:"userId"`
	RoleID        int    `json:"roleId"`
}

// Login 模拟登录
func Login(username, password string) (token string, userID int, err error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	data := url.Values{
		"loginName": {username},
		"password":  {password},
	}
	var req *http.Request
	req, err = http.NewRequest("POST", global.Config.Ulearning.Login, strings.NewReader(data.Encode()))
	if err != nil {
		zlog.Errorf("请求登录失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		zlog.Errorf("请求登录失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := ioutil.ReadAll(resp.Body)
		zlog.Errorf("请求登录失败: %s", string(body))
		return
	}

	authToken := ""
	userInfo := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "AUTHORIZATION" {
			authToken = cookie.Value
		} else if cookie.Name == "USERINFO" {
			userInfo, err = url.QueryUnescape(cookie.Value)
			if err != nil {
				zlog.Errorf("解析用户信息失败: %v", err)
				return
			}
		}
	}

	if authToken == "" {
		zlog.Errorf("登录失败: 未获取到 AUTHORIZATION 值")
		return
	}

	var userDetails LoginResponse
	if err = json.Unmarshal([]byte(userInfo), &userDetails); err != nil {
		zlog.Errorf("解析用户信息失败: %v", err)
		return
	}

	//fmt.Printf("UserID: %s, RoleID: %d\n", userDetails.UserID, userDetails.RoleID)

	token = authToken
	userID = userDetails.UserID

	zlog.Debugf("登录成功 %v", userInfo)
	return
}
