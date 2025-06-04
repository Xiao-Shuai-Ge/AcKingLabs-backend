package logic

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/ulearning"
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
	user := ulearning.NewUser()
	user.Login(userInfo.UserName, userInfo.Password)
	zlog.CtxDebugf(ctx, "模拟用户登录成功: token=%s, userID=%d", user.Token, user.UserID)
	teacher := ulearning.NewUser()
	teacher.TeacherLogin()
	zlog.CtxDebugf(ctx, "模拟老师登录成功: token=%s, userID=%d", teacher.Token, teacher.UserID)

	// 获取课程列表
	Courses, err := user.GetAllCourses()
	if err != nil {
		zlog.Errorf("获取课程列表失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	for _, course := range Courses.CourseList {
		// 填装返回数据
		resp.Courses = append(resp.Courses, types.Course{
			CourseID:   course.ID,
			CourseName: course.Name,
			ClassID:    course.ClassID,
		})
		// 获取课程活动
		Activities, err := user.GetCourseActivities(course.ID)
		if err != nil {
			zlog.Errorf("获取课程活动失败: %v", err)
			return resp, response.ErrResp(err, response.INTERNAL_ERROR)
		}
		for _, activity := range Activities.OtherActivityDTOList {
			//if activity.RelationType == 1 {
			//	zlog.CtxDebugf(ctx, "课程name:%s 状态:%d", course.Name, activity.PersonStatus)
			//}
			if activity.RelationType == 1 && activity.PersonStatus != 1 {
				zlog.CtxDebugf(ctx, "课程活动名称: %s", activity.Title)
				// 获取详细数据
				detail, err := teacher.GetActivityDetail(activity.RelationID)
				if err != nil {
					zlog.CtxErrorf(ctx, "获取课程活动详细信息失败: %v", err)
					return resp, response.ErrResp(err, response.INTERNAL_ERROR)
				}
				zlog.CtxDebugf(ctx, "课程活动详细信息: %v", detail)
				// 填装返回数据
				Active := types.Active{
					UserID:        user.UserID,
					ClassName:     course.Name,
					ActiveName:    activity.Title,
					ClassID:       course.ClassID,
					RelationID:    activity.RelationID,
					AbsenceNum:    detail.AbsenceNum,
					NotAbsenceNum: detail.NotAbsenceNum,
					IsFinished:    detail.Finish == "0",
				}
				resp.Actives = append(resp.Actives, Active)
			}
		}
	}
	resp.Length = len(resp.Actives)
	resp.Level = userInfo.Level
	zlog.CtxDebugf(ctx, "签到列表: %v", resp)
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
	user := ulearning.NewUser()
	err = user.Login(userInfo.UserName, userInfo.Password)
	if err != nil {
		zlog.CtxErrorf(ctx, "模拟登录失败: %v", err)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	zlog.CtxDebugf(ctx, "模拟登录成功: token=%s, userID=%d", user.Token, user.UserID)
	// 刷新教师 token
	err = ulearning.NewUser().TeacherLogin()
	if err != nil {
		zlog.CtxErrorf(ctx, "刷新教师 token 失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	// 签到
	err = user.SigninByStudent(req.RelationID, req.ClassID)
	if err != nil {
		zlog.CtxErrorf(ctx, "签到失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	resp.Message = "签到成功"
	return
}

func (l *TemplateLogic) SigninTeacher(ctx context.Context, req types.SigninTeacherReq) (resp types.SigninTeacherResp, err error) {
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
	} else if userInfo.Level < 1 {
		zlog.CtxErrorf(ctx, "用户 %d 权限不足", userID)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	zlog.CtxDebugf(ctx, "用户 %d 权限为 %d", userID, userInfo.Level)
	// 模拟登录
	user := ulearning.NewUser()
	err = user.Login(userInfo.UserName, userInfo.Password)
	if err != nil {
		zlog.CtxErrorf(ctx, "模拟登录失败: %v", err)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	zlog.CtxDebugf(ctx, "模拟登录成功: token=%s, userID=%d", user.Token, user.UserID)
	// 登录老师账号
	teacher := ulearning.NewUser()
	err = teacher.TeacherLogin()
	if err != nil {
		zlog.CtxErrorf(ctx, "模拟老师登录失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	// 签到
	err = teacher.SigninByTeacher(req.RelationID, user.UserID)
	if err != nil {
		zlog.CtxErrorf(ctx, "签到失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	resp.Message = "签到成功"
	return
}

func (l *TemplateLogic) GetAutoList(ctx context.Context, req types.GetAutoListReq) (resp types.GetAutoListResp, err error) {
	defer utils.RecordTime(time.Now())()
	// id 转 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库获取自动签到列表
	autoList, err := repo.NewTemplateRepo(global.DB).GetAutoSigninListByUserID(userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取自动签到列表失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	for _, auto := range autoList {
		// 填装返回数据
		resp.CourseList = append(resp.CourseList, types.AutoCourse{
			CourseID: int(auto.CoursesID),
			Percent:  auto.Percentage,
		})
	}
	return
}

func (l *TemplateLogic) AutoSetting(ctx context.Context, req types.AutoSettingReq) (resp types.AutoSettingResp, err error) {
	defer utils.RecordTime(time.Now())()
	// id 转 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库拿去用户权限
	userInfo, err := repo.NewTemplateRepo(global.DB).GetUserInfo(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户 %d 不在白名单中，权限不足", userID)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	} else if userInfo.Level < 3 {
		zlog.CtxErrorf(ctx, "用户 %d 权限不足", userID)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 填装数据
	// 拿取邮箱信息
	user, err := repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取用户邮箱信息失败: %v", err)
		return resp, response.ErrResp(err, response.INTERNAL_ERROR)
	}
	email := user.Email
	auto := model.AutoSignin{
		UserID:     userID,
		CoursesID:  int64(req.CourseID),
		CourseName: req.CourseName,
		Percentage: req.Percent,
		ClassID:    int64(req.ClassID),
		Email:      email,
	}
	// 更改设置
	if req.IsAuto {
		// 判断范围
		if req.Percent < 0 || req.Percent > 100 {
			zlog.CtxErrorf(ctx, "签到百分比 %d 超出范围", req.Percent)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		// 判断是否已经存在
		var isExist bool
		isExist, err = repo.NewTemplateRepo(global.DB).IsExistAuto(userID, int64(req.CourseID))
		if err != nil {
			zlog.CtxErrorf(ctx, "获取自动签到设置是否存在失败: %v", err)
			return resp, response.ErrResp(err, response.INTERNAL_ERROR)
		}
		if isExist {
			// 更新
			zlog.CtxDebugf(ctx, "更新自动签到设置: %v", auto)
			err = repo.NewTemplateRepo(global.DB).UpdateAutoSignin(&auto)
		} else {
			// 添加设置
			err = repo.NewTemplateRepo(global.DB).CreateAutoSignin(&auto)
		}
	} else {
		// 取消设置
		err = repo.NewTemplateRepo(global.DB).DeleteAutoSignin(&auto)
	}
	resp.IsAuto = req.IsAuto
	return
}
