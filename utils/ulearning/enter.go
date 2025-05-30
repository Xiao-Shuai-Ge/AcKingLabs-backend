package ulearning

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
)

type User struct {
	Token  string
	UserID int
}

var teacherUser *User

func NewUser() *User {
	return &User{}
}

// TeacherLogin 默认教师登录方法
func (l *User) TeacherLogin() (err error) {
	err = l.Login(global.Config.Ulearning.Teacher, global.Config.Ulearning.Password)
	teacherUser = l
	return
}

// RefreshTeacherToken 刷新教师 Token
func RefreshTeacherToken() (err error) {
	teacherUser = NewUser()
	err = teacherUser.TeacherLogin()
	return
}

// Login 基础登录方法
func (l *User) Login(username, password string) (err error) {
	// 构建请求参数
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

	var userDetails LoginResp
	if err = json.Unmarshal([]byte(userInfo), &userDetails); err != nil {
		zlog.Errorf("解析用户信息失败: %v", err)
		return
	}

	l.Token = authToken
	l.UserID = userDetails.UserID
	zlog.Debugf("Token: %s, UserID: %d", l.Token, l.UserID)
	return
}

// GetAllCourses 获取所有课程列表
func (l *User) GetAllCourses() (resp GetAllCoursesResp, err error) {
	// 构建请求参数
	Url := global.Config.Ulearning.GetAllCourses
	geq := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded", // 明确设置 Content-Type
			"Authorization": l.Token,
		},
	}
	var response *grequests.Response
	response, err = grequests.Get(Url, geq)
	if err != nil {
		zlog.Errorf("请求课程列表失败: %v", err)
		return
	}
	defer response.Close()
	// 解析响应数据
	if err = response.JSON(&resp); err != nil {
		zlog.Errorf("解析响应失败: %v", err)
		return
	}
	//zlog.Debugf("课程列表: %v", resp)
	return
}

// GetCourseActivities 获取课程活动列表
func (l *User) GetCourseActivities(courseID int) (resp GetCourseActivitiesResp, err error) {
	// 构建请求参数
	Url := fmt.Sprintf(global.Config.Ulearning.GetCourseActivities, courseID)
	geq := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded", // 明确设置 Content-Type
			"Authorization": l.Token,
		},
	}
	var response *grequests.Response
	response, err = grequests.Get(Url, geq)
	if err != nil {
		zlog.Errorf("请求课程活动失败: %v", err)
		return
	}
	defer response.Close()
	// 解析响应数据
	if err = response.JSON(&resp); err != nil {
		zlog.Errorf("解析响应失败: %v", err)
		return
	}
	//zlog.Debugf("课程活动: %v", resp)
	return
}

// GetActivityDetail 获取活动详情
func (l *User) GetActivityDetail(relationID int) (resp GetActivityDetailResp, err error) {
	Url := fmt.Sprintf(global.Config.Ulearning.GetActivityDetail, relationID)
	geq := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded", // 明确设置 Content-Type
			"Authorization": l.Token,
		},
	}
	var response *grequests.Response
	response, err = grequests.Get(Url, geq)
	if err != nil {
		zlog.Errorf("请求活动详情失败: %v", err)
		return
	}
	defer response.Close()
	// 解析响应数据
	if err = response.JSON(&resp); err != nil {
		zlog.Errorf("解析响应失败: %v", err)
		return
	}
	zlog.Debugf("活动详情: %v", resp)
	return
}

// SigninByTeacher 老师签到
func (l *User) SigninByTeacher(relationID int, userID int) (err error) {
	user := SigninUser{
		UserID: userID,
		Status: 1,
	}
	users := []SigninUser{user}
	var reqData = SigninTeacherOPReq{
		AttendanceID: relationID,
		Users:        users,
	}
	Url := global.Config.Ulearning.SigninTeacherOperation
	geq := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": l.Token,
		},
		JSON: reqData,
	}
	var response *grequests.Response
	response, err = grequests.Post(Url, geq)
	if err != nil {
		zlog.Errorf("请求签到失败: %v", err)
		return
	}
	defer response.Close()
	// 解析响应数据
	var resp SigninTeacherOPResp
	if err = response.JSON(&resp); err != nil {
		zlog.Errorf("解析响应失败: %v", err)
		return
	}
	if resp.Status != "success" {
		zlog.Errorf("签到失败: %s", resp.Msg)
		return errors.New(resp.Msg)
	}
	zlog.Infof("签到成功")
	return
}

func (l *User) SigninByStudent(relationID int, classID int) (err error) {
	// 先获取签到信息
	data, err := teacherUser.GetActivityDetail(relationID)
	// 构建请求参数
	Url := global.Config.Ulearning.SigninOperation
	var reqData = SigninOperationReq{
		AttendanceID:   relationID,
		ClassID:        classID,
		UserID:         l.UserID,
		Location:       data.Location,
		Address:        "",
		EnterWay:       1,
		AttendanceCode: data.AttendanceCode,
	}
	geq := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": l.Token,
		},
		JSON: reqData,
	}
	var response *grequests.Response
	response, err = grequests.Post(Url, geq)
	if err != nil {
		zlog.Errorf("请求签到失败: %v", err)
		return
	}
	defer response.Close()
	// 解析响应数据
	var resp SigninOperationResp
	if err = response.JSON(&resp); err != nil {
		zlog.Errorf("解析响应失败: %v", err)
		return
	}
	if resp.NewStatus != 1 {
		zlog.Errorf("签到失败: %s", resp.Msg)
		return errors.New(resp.Msg)
	}
	zlog.Infof("签到成功")
	return
}
