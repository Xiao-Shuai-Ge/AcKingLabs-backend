package initalize

import (
	"runtime"
	"tgwp/global"
	"tgwp/internal/utils/task"
	"tgwp/log/zlog"
)

func Eve() {
	zlog.Warnf("开始释放资源！")
	if task.GlobalDispatcher != nil {
		task.GlobalDispatcher.Stop()
	}
	errRedis := global.Rdb.Close()
	if errRedis != nil {
		zlog.Errorf("Redis关闭失败 ：%v", errRedis.Error())
	}
	sqlDB, _ := global.DB.DB()
	errDB := sqlDB.Close()
	if errDB != nil {
		zlog.Errorf("数据库关闭失败 ：%v", errDB.Error())
	}
	runtime.GC()
	if errDB == nil && errRedis == nil {
		zlog.Warnf("资源释放成功！")
	}
}
