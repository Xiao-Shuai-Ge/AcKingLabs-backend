package logic

import (
	"context"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/email"
	"tgwp/utils/jwtUtils"
	"time"
)

type LoginLogic struct {
}

const (
	EMAIL_REGEX      = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	REDIS_EMAIL_CODE = "login:email:%s:code"
)

func NewLoginLogic() *LoginLogic {
	return &LoginLogic{}
}

// SendCode 发送验证码
func (l *LoginLogic) SendCode(ctx context.Context, req types.SendCodeReq) (resp types.SendCodeResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 验证邮箱格式
	re := regexp.MustCompile(EMAIL_REGEX, 0)
	if isMatch, _ := re.MatchString(req.Email); !isMatch {
		return resp, response.ErrResp(err, response.EMAIL_NOT_VALID)
	}
	// 生成随机验证码
	code := rand.Intn(1000000)
	zlog.CtxDebugf(ctx, "生成验证码: %d", code)
	// 保存验证码到redis
	err = global.Rdb.Set(ctx, fmt.Sprintf(REDIS_EMAIL_CODE, req.Email), code, 5*time.Minute).Err()
	if err != nil {
		return resp, response.ErrResp(err, response.REDIS_ERROR)
	}
	// 发送验证码
	err = email.SendCode(req.Email, int64(code))
	if err != nil {
		return resp, response.ErrResp(err, response.EMAIL_SEND_ERROR)
	}
	// 发送邮箱成功
	zlog.CtxDebugf(ctx, "发送邮箱成功: %v", req)
	return resp, nil
}

// Register 注册
func (l *LoginLogic) Register(ctx context.Context, req types.RegisterReq) (resp types.RegisterResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 验证用户名格式
	if len(req.Username) > 30 {
		zlog.CtxInfof(ctx, "用户名格式错误: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 验证密码格式
	if len(req.Password) > 30 {
		zlog.CtxInfof(ctx, "密码格式错误: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 验证邮箱格式
	re := regexp.MustCompile(EMAIL_REGEX, 0)
	if isMatch, _ := re.MatchString(req.Email); !isMatch {
		zlog.CtxInfof(ctx, "邮箱格式错误: %v", err)
		return resp, response.ErrResp(err, response.EMAIL_NOT_VALID)
	}
	// 验证验证码
	code, err := global.Rdb.Get(ctx, fmt.Sprintf(REDIS_EMAIL_CODE, req.Email)).Int()
	if err != nil {
		// 如果Redis里没有验证码，说明验证码过期或者压根没发送过
		zlog.CtxInfof(ctx, "验证码错误: %v", err)
		return resp, response.ErrResp(err, response.VERIFY_CODE_VALID)
	}
	if fmt.Sprintf("%06d", code) != req.Code {
		zlog.CtxInfof(ctx, "验证码错误: %v", err)
		return resp, response.ErrResp(err, response.VERIFY_CODE_VALID)
	}
	// 满足条件，创建用户
	// 密码加密
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	zlog.CtxDebugf(ctx, "密码加密成功: %v", string(HashPassword))
	if err != nil {
		zlog.CtxInfof(ctx, "密码加密失败: %v", err)
		return resp, response.ErrResp(err, response.VERIFY_CODE_VALID)
	}
	// 创建用户
	id := global.SnowflakeNode.Generate().Int64()
	user := model.User{
		ID:       id,
		Username: req.Username,
		Password: string(HashPassword),
		Email:    req.Email,
	}
	// 放入数据库
	err = repo.NewLoginRepo(global.DB).AddUser(user)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建用户失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 生成 atoken
	atoken, err := jwtUtils.GenAtoken(fmt.Sprintf("%d", id), req.Username, global.ATOKEN_EFFECTIVE_TIME)
	resp.Atoken = atoken
	return resp, nil
}

func (l *LoginLogic) TokenTest(ctx context.Context, req types.TokenTestReq) (resp types.TokenTestResp, err error) {
	return
}
