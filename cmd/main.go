package main

import (
	"tgwp/initalize"
	"tgwp/log/zlog"
	routerg "tgwp/router"
)

func main() {
	// 初始化
	initalize.Init()

	// 工程进入前夕，释放资源
	defer initalize.Eve()

	// 运行服务
	routerg.RunServer()
	zlog.Infof("程序运行完成！")
}
