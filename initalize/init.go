package initalize

import (
	"tgwp/cmd/flags"
	"tgwp/global"
	"tgwp/utils"
)

func Init() {
	// 解析命令行参数
	flags.Parse()

	// 启动前缀展示
	introduce()

	// 初始化根目录
	InitPath()

	// 加载配置文件
	InitConfig()

	// 正式初始化日志
	InitLog(global.Config)

	// 初始化数据库
	InitDataBase(*global.Config)
	InitRedis(*global.Config)

	// 对命令行参数进行处理
	flags.Run()
}

func InitPath() {
	global.Path = utils.GetRootPath("")
}
