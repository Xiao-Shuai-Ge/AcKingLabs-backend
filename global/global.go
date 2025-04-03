package global

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"tgwp/configs"
)

var (
	Path   string
	DB     *gorm.DB
	Rdb    *redis.Client
	Config *configs.Config
)
