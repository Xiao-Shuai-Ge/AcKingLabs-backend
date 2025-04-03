package response

type MsgCode struct {
	Code int
	Msg  string
}

//五位业务状态码

var (
	/* 成功 */
	SUCCESS = MsgCode{Code: 20000, Msg: "成功"}

	/* 默认失败 */
	COMMON_FAIL = MsgCode{-43960, "失败"}

	/* 请求错误 <0 */
	TOKEN_IS_EXPIRED = MsgCode{-20000, "token已过期"}
	TOKEN_IS_BLANK   = MsgCode{-20001, "token为空"}
	TOKEN_NOT_VALID  = MsgCode{-20002, "token无效"}
	TOKEN_TYPE_ERROR = MsgCode{-20003, "token类型错误"}

	/* 内部错误 60000 ~ 69999 */
	INTERNAL_ERROR              = MsgCode{60001, "内部错误, check log"}
	INTERNAL_FILE_UPLOAD_ERROR  = MsgCode{60002, "文件上传失败"}
	SNOWFLAKE_ID_GENERATE_ERROR = MsgCode{60003, "snowflake id生成失败"}
	DATABASE_ERROR              = MsgCode{60004, "数据库错误"}
	REDIS_ERROR                 = MsgCode{60005, "redis错误"}
	FEISHU_ERROR                = MsgCode{60006, "飞书API错误"}

	/* 参数错误：10000 ~ 19999 */
	PARAM_NOT_VALID    = MsgCode{10001, "参数无效"}
	PARAM_IS_BLANK     = MsgCode{10002, "参数为空"}
	PARAM_TYPE_ERROR   = MsgCode{10003, "参数类型错误"}
	PARAM_NOT_COMPLETE = MsgCode{10004, "参数缺失"}
	MEMBER_NOT_EXIST   = MsgCode{10005, "用户不存在"}
	MESSAGE_NOT_EXIST  = MsgCode{10006, "消息不存在"}

	/* 用户错误 20000 ~ 29999 */
	USER_NOT_LOGIN             = MsgCode{20001, "用户未登录"}
	USER_PASSWORD_DIFFERENT    = MsgCode{20002, "用户两次密码输入不一致"}
	USER_ACCOUNT_NOT_EXIST     = MsgCode{20003, "账号不存在"}
	USER_CREDENTIALS_ERROR     = MsgCode{20004, "密码错误"}
	PHONE_ERROR                = MsgCode{20005, "手机号格式错误"}
	CONNECT_PHONE_ERROR        = MsgCode{20006, "连接不到用户手机"}
	USER_ACCOUNT_ALREADY_EXIST = MsgCode{20007, "账号已存在"}
	CAPTCHA_ERROR              = MsgCode{20500, "验证码错误"}
	INSUFFICENT_PERMISSIONS    = MsgCode{20403, "权限不足"}
	FEISHU_OPENID_NOT_EXIST    = MsgCode{20008, "该用户手机号未绑定飞书"}

	/*
	 USER_ACCOUNT_DISABLE(20005, "账号不可用"),
	 USER_ACCOUNT_LOCKED(20006, "账号被锁定"),
	 USER_ACCOUNT_NOT_EXIST(20007, "账号不存在"),
	 USER_ACCOUNT_USE_BY_OTHERS(20009, "账号下线"),
	 USER_ACCOUNT_EXPIRED(20010, "账号已过期"),
	*/
)
