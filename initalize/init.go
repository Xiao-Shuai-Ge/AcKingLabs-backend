package initalize

import (
	"fmt"
	"tgwp/cmd/flags"
	"tgwp/global"
	"tgwp/internal/utils/messageService"
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

	fmt.Println(global.Config.DB.Dsn)
	// 正式初始化日志
	InitLog(global.Config)

	// 初始化数据库
	InitDataBase(*global.Config)
	InitRedis(*global.Config)

	// 初始化全局雪花ID生成器
	InitSnowflake()

	// 开启定时任务
	Cron()

	// 初始化OSS服务
	InitOSS()

	// 初始化ElasticSearch
	InitElasticsearch()

	// 启动消息服务
	messageService.GetMessageService().Start()

	// 对命令行参数进行处理
	flags.Run()

	// 测试AI聊天
	//logic.AiChat()

}

func InitPath() {
	global.Path = utils.GetRootPath("")
}
