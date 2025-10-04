package global

import (
	"tgwp/utils/snowflake"
	"time"
)

// 所有常量文件读取位置
const (
	DEFAULT_CONFIG_FILE_PATH = "/config.yaml"
	ATOKEN_EFFECTIVE_TIME    = time.Hour * 2
	//ATOKEN_EFFECTIVE_TIME = time.Second * 10

	RTOKEN_EFFECTIVE_TIME = time.Hour * 24 * 7 * 2
	AUTH_ENUMS_ATOKEN     = "atoken"
	AUTH_ENUMS_RTOKEN     = "rtoken"
	DEFAULT_NODE_ID       = 1
	TOKEN_USER_ID         = "user_id"
	TOKEN_ROLE            = "role"

	ROLE_NOT_LOGIN   = -1
	ROLE_GUEST       = 0
	ROLE_USER        = 1
	ROLE_PLAYER      = 2
	ROLE_ADMIN       = 3
	ROLE_SUPER_ADMIN = 4

	// 帖子内容长度限制（字符数）
	POST_MAX_LENGTH_USER   = 15000 // 普通用户
	POST_MAX_LENGTH_PLAYER = 30000 // 正式选手
	POST_MAX_LENGTH_ADMIN  = 50000 // 管理员及以上

	// 评论内容长度限制（字符数）
	COMMENT_MAX_LENGTH_USER   = 1000 // 普通用户
	COMMENT_MAX_LENGTH_PLAYER = 2000 // 正式选手
	COMMENT_MAX_LENGTH_ADMIN  = 5000 // 管理员及以上
)

var (
	TYPE_SET = map[string]bool{
		"diary":    true,
		"tutorial": true,
		"solution": true,
		"contest":  true,
		"fun":      true,
		"official": true,
		"help":     true,
	}
)

var Node, _ = snowflake.NewNode(DEFAULT_NODE_ID)
