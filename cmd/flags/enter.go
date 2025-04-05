package flags

import (
	"flag"
	"fmt"
	"os"
	"tgwp/global"
	"tgwp/model"
)

type Options struct {
	File    string
	DB      bool
	Version bool
}

var FlagOptions = new(Options)

func Parse() {
	flag.StringVar(&FlagOptions.File, "f", "config.yaml", "配置文件")
	flag.BoolVar(&FlagOptions.DB, "db", false, "数据库迁移")
	flag.BoolVar(&FlagOptions.Version, "v", false, "版本信息")
	flag.Parse()
}
func Run() {
	//为了不每次启动都迁移表，但要迁移表，可以手动执行go run cmd/main.go -db
	if FlagOptions.DB {
		//执行数据库迁移
		migrateTables()
		os.Exit(0)
	}
}
func migrateTables() {
	//自动迁移某一个表，确保表结构存在
	err := global.DB.AutoMigrate(
		&model.User{},
	)
	if err != nil {
		fmt.Println("数据库迁移失败！")
	}
	fmt.Println("数据库迁移成功！")
}
