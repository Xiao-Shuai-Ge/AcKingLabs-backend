package global

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"tgwp/configs"
	"tgwp/utils/snowflake"
)

var (
	Path   string
	DB     *gorm.DB
	Rdb    *redis.Client
	Config *configs.Config

	SnowflakeNode *snowflake.Node // 默认雪花ID生成节点
)
