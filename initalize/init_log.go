package initalize

import (
	"tgwp/configs"
	"tgwp/log"
	"tgwp/log/zlog"
)

func InitLog(config *configs.Config) {
	logger := log.GetZap(config)
	zlog.InitLogger(logger)
}
