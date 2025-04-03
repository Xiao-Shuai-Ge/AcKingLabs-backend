package mysqlx

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"tgwp/configs"
	"tgwp/log/zlog"
	"tgwp/pkg/database"
)

type Mysql struct {
}

// InitDataBases 初始化
func (m *Mysql) InitDataBase(config configs.Config) (*gorm.DB, error) {
	dsn := m.GetDsn(config)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		zlog.Panicf("MySQL无法连接数据库！: %v", err)
		return nil, err
	}
	zlog.Infof("MySQL连接数据库成功！")
	return db, nil
}
func (m *Mysql) GetDsn(config configs.Config) string {
	return config.DB.Dsn
}
func NewMySql() database.DataBase {
	return &Mysql{}
}
