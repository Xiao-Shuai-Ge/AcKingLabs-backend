package initalize

import (
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/utils/snowflake"
)

func InitSnowflake() {
	var err error
	global.SnowflakeNode, err = snowflake.NewNode(global.DEFAULT_NODE_ID)
	if err != nil {
		zlog.Errorf("初始化雪花ID生成节点失败: %v", err)
		panic(err)
	}
}
