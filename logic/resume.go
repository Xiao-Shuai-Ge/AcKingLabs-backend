package logic

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/email"
	"time"

	"gorm.io/gorm"
)

type ResumeLogic struct {
}

func NewResumeLogic() *ResumeLogic {
	return &ResumeLogic{}
}

// SubmitResume 投递简历
func (l *ResumeLogic) SubmitResume(ctx context.Context, req types.SubmitResumeReq) (resp types.SubmitResumeResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "投递简历请求: %v", req)

	// 验证邮箱验证码
	err = l.verifyEmailCode(ctx, req.Email, req.Code)
	if err != nil {
		return resp, err
	}

	// 检查邮箱是否已经投递过简历
	_, err = repo.NewResumeRepo(global.DB).GetResumeByEmail(req.Email)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "该邮箱已经投递过简历: %s", req.Email)
		return resp, response.ErrResp(err, response.RESUME_ALREADY_EXIST)
	}

	// 验证额外信息字段
	err = l.validateExtraFields(req.Extra)
	if err != nil {
		return resp, err
	}

	// 创建简历
	resume := model.Resume{
		RealName:   strings.TrimSpace(req.RealName),
		Grade:      req.Grade,
		StudentNo:  strings.TrimSpace(req.StudentNo),
		Email:      strings.TrimSpace(req.Email),
		Extra:      req.Extra,
		IsAccepted: false,
	}

	err = repo.NewResumeRepo(global.DB).CreateResume(resume)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 获取创建的简历ID
	createdResume, err := repo.NewResumeRepo(global.DB).GetResumeByEmail(req.Email)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取创建的简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	resp.ID = createdResume.ID
	zlog.CtxInfof(ctx, "投递简历成功: %d", resp.ID)
	return resp, nil
}

// UpdateResume 修改简历
func (l *ResumeLogic) UpdateResume(ctx context.Context, req types.UpdateResumeReq) (resp types.UpdateResumeResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "修改简历请求: %v", req)

	// 验证邮箱验证码
	err = l.verifyEmailCode(ctx, req.Email, req.Code)
	if err != nil {
		return resp, err
	}

	// ID 转化为 int64
	resumeID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%s 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 获取简历
	resume, err := repo.NewResumeRepo(global.DB).GetResumeByID(resumeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "简历不存在: %v", err)
		return resp, response.ErrResp(err, response.RESUME_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 验证邮箱是否匹配
	if resume.Email != strings.TrimSpace(req.Email) {
		zlog.CtxErrorf(ctx, "邮箱不匹配")
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}

	// 验证额外信息字段
	err = l.validateExtraFields(req.Extra)
	if err != nil {
		return resp, err
	}

	// 更新简历信息
	resume.RealName = strings.TrimSpace(req.RealName)
	resume.Grade = req.Grade
	resume.StudentNo = strings.TrimSpace(req.StudentNo)
	resume.Email = strings.TrimSpace(req.Email)
	resume.Extra = req.Extra

	err = repo.NewResumeRepo(global.DB).UpdateResume(resume)
	if err != nil {
		zlog.CtxErrorf(ctx, "更新简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	zlog.CtxInfof(ctx, "修改简历成功: %d", resumeID)
	return resp, nil
}

// GetResumeDetail 查询简历详细信息
func (l *ResumeLogic) GetResumeDetail(ctx context.Context, req types.GetResumeDetailReq, userRole int) (resp types.GetResumeDetailResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "查询简历详细信息请求: %v", req)

	// ID 转化为 int64
	resumeID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%s 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 获取简历
	resume, err := repo.NewResumeRepo(global.DB).GetResumeByID(resumeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "简历不存在: %v", err)
		return resp, response.ErrResp(err, response.RESUME_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 如果非管理员，需要有邮箱验证码才能通过
	if userRole < 3 {
		if l.verifyEmailCode(ctx, resume.Email, resume.Code) != nil {
			zlog.CtxErrorf(ctx, "权限不足，需要管理员权限")
			return resp, response.ErrResp(err, response.PERMISSION_DENIED)
		}
	}

	// 填入响应数据
	resp.ID = resume.ID
	resp.RealName = resume.RealName
	resp.Grade = resume.Grade
	resp.StudentNo = resume.StudentNo
	resp.Email = resume.Email
	resp.Extra = resume.Extra
	resp.Code = resume.Code
	resp.IsAccepted = resume.IsAccepted
	resp.CreatedAt = time.UnixMilli(resume.CreatedTime).Format("2006-01-02 15:04:05")
	resp.UpdatedAt = time.UnixMilli(resume.UpdatedTime).Format("2006-01-02 15:04:05")

	zlog.CtxInfof(ctx, "查询简历详细信息成功: %d", resumeID)
	return resp, nil
}

// GetResumeList 获取简历列表（管理员功能）
func (l *ResumeLogic) GetResumeList(ctx context.Context, req types.GetResumeListReq) (resp types.GetResumeListResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "获取简历列表请求: %v", req)

	// 获取简历列表
	resumes, total, err := repo.NewResumeRepo(global.DB).GetResumeList(req.Page, req.Count)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取简历列表失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 填入参数
	for _, resume := range resumes {
		resp.Resumes = append(resp.Resumes, types.ResumeListItem{
			ID:         resume.ID,
			RealName:   resume.RealName,
			Grade:      resume.Grade,
			StudentNo:  resume.StudentNo,
			Email:      resume.Email,
			IsAccepted: resume.IsAccepted,
			CreatedAt:  time.UnixMilli(resume.CreatedTime).Format("2006-01-02 15:04:05"),
			UpdatedAt:  time.UnixMilli(resume.UpdatedTime).Format("2006-01-02 15:04:05"),
		})
	}

	// 计算总页数
	if int(total)%req.Count == 0 {
		resp.PageTotal = int64(int(total) / req.Count)
	} else {
		resp.PageTotal = int64(int(total)/req.Count + 1)
	}

	resp.Length = len(resumes)
	resp.Total = total

	zlog.CtxInfof(ctx, "获取简历列表成功，共 %d 条记录", len(resumes))
	return resp, nil
}

// DeleteResume 删除简历（管理员功能）
func (l *ResumeLogic) DeleteResume(ctx context.Context, req types.DeleteResumeReq) (resp types.DeleteResumeResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "删除简历请求: %v", req)

	// ID 转化为 int64
	resumeID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%s 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 检查简历是否存在
	_, err = repo.NewResumeRepo(global.DB).GetResumeByID(resumeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "简历不存在: %v", err)
		return resp, response.ErrResp(err, response.RESUME_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 删除简历
	err = repo.NewResumeRepo(global.DB).DeleteResume(resumeID)
	if err != nil {
		zlog.CtxErrorf(ctx, "删除简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	zlog.CtxInfof(ctx, "删除简历成功: %d", resumeID)
	return resp, nil
}

// AcceptResume 通过简历（管理员功能）
func (l *ResumeLogic) AcceptResume(ctx context.Context, req types.AcceptResumeReq) (resp types.AcceptResumeResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
	zlog.CtxInfof(ctx, "通过简历请求: %v", req)

	// ID 转化为 int64
	resumeID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%s 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 获取简历
	resume, err := repo.NewResumeRepo(global.DB).GetResumeByID(resumeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "简历不存在: %v", err)
		return resp, response.ErrResp(err, response.RESUME_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 检查简历是否已经被通过
	if resume.IsAccepted {
		zlog.CtxErrorf(ctx, "简历已经被通过: %d", resumeID)
		return resp, response.ErrResp(err, response.RESUME_ALREADY_ACCEPTED)
	}

	// 生成6位随机大写字母邀请码
	code := l.generateInvitationCode()

	// 发送邀请码邮件
	err = email.SendInvitationCodeEmail(resume.Email, code)
	if err != nil {
		zlog.CtxErrorf(ctx, "发送邀请码邮件失败: %v", err)
		// 邮件发送失败也中断简历通过流程，以防对方无法收到邀请码
	}

	// 通过简历
	err = repo.NewResumeRepo(global.DB).AcceptResume(resumeID, code)
	if err != nil {
		zlog.CtxErrorf(ctx, "通过简历失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	resp.Code = code
	zlog.CtxInfof(ctx, "通过简历成功: %d, 邀请码: %s", resumeID, code)
	return resp, nil
}

// verifyEmailCode 验证邮箱验证码
func (l *ResumeLogic) verifyEmailCode(ctx context.Context, emailAddr, code string) error {
	// 从Redis获取验证码
	storedCode, err := global.Rdb.Get(ctx, fmt.Sprintf(REDIS_EMAIL_CODE, emailAddr)).Int()
	if err != nil {
		zlog.CtxErrorf(ctx, "验证码不存在或已过期: %v", err)
		return response.ErrResp(err, response.VERIFY_CODE_VALID)
	}

	// 验证验证码
	if fmt.Sprintf("%06d", storedCode) != code {
		zlog.CtxErrorf(ctx, "验证码错误")
		return response.ErrResp(err, response.VERIFY_CODE_VALID)
	}

	// 验证码可以重复使用，以便于获取最新简历后，重新发送验证码
	// global.Rdb.Del(ctx, fmt.Sprintf(REDIS_EMAIL_CODE, emailAddr))
	return nil
}

// validateExtraFields 验证额外信息字段
func (l *ResumeLogic) validateExtraFields(extra map[string]string) error {
	requiredFields := []string{"information", "skills", "reason", "understanding", "future_plan"}

	for _, field := range requiredFields {
		if value, exists := extra[field]; !exists || strings.TrimSpace(value) == "" {
			return response.ErrResp(nil, response.PARAM_NOT_VALID)
		}
	}

	return nil
}

// generateInvitationCode 生成6位随机大写字母邀请码
func (l *ResumeLogic) generateInvitationCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}
