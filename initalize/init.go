package initalize

import (
	"fmt"
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

	// 对命令行参数进行处理
	flags.Run()

	//user := ulearning.NewUser()

	//user.Login("dgut2023463030521", "Yxy123456")

	//user.GetAllCourses()
	//user.GetCourseActivities(151267)
	//teacher.SigninByTeacher(842804, user.UserID)
	//user.SigninByStudent(842804, 881864)

	//BookingContestNotice()
	//email.BookingContest([]string{"1019513201@qq.com", "1228341152@qq.com"}, "https://oj.dgut.edu.cn/", "2077年 东莞理工大学 软通杯", "20")
	//zlog.Debugf("爬取比赛内容: %v", contest.GetNowcoderContest())

}

func InitPath() {
	global.Path = utils.GetRootPath("")
}
