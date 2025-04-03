package blacklist

import (
	"context"
	"fmt"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/utils/jwtUtils"
	"time"
)

type BlackType int8

const (
	UserBlackType   BlackType = iota + 1 //用户自己注销
	AdminBlackType                       //管理员删除账号
	DeviceBlackType                      //其它设备登录挤下线
)

func (b BlackType) String() string {
	return fmt.Sprintf("%d", b)
}
func (b BlackType) ParseBlackType(val string) BlackType {
	switch val {
	case "1":
		return UserBlackType
	case "2":
		return AdminBlackType
	case "3":
		return DeviceBlackType
	}
	return UserBlackType
}
func (b BlackType) Msg() string {
	switch b {
	case UserBlackType:
		return "已注销"
	case AdminBlackType:
		return "禁止登录"
	case DeviceBlackType:
		return "设备已下线"
	}
	return "已注销"
}
func TokenBlack(ctx context.Context, token string, value BlackType) error {
	key := fmt.Sprintf("token_black_%s", token)
	claims, err := jwtUtils.ParseToken(token)
	if err != nil || claims == nil {
		return err
	}
	global.Rdb.Get(ctx, key)
	t := time.Duration(claims.RegisteredClaims.ExpiresAt.Unix() - time.Now().Unix())
	_, err = global.Rdb.Set(ctx, key, value.String(), t*time.Second).Result()
	if err != nil {
		zlog.Errorf("redis存放黑名单错误:%v", err)
		return err
	}
	return nil
}
func HasTokenBlack(ctx context.Context, token string) (value BlackType, exist bool, err error) {
	key := fmt.Sprintf("token_black_%s", token)
	res, err := global.Rdb.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			// 黑名单不存在对象
			return value, false, nil
		}
		zlog.CtxErrorf(ctx, "redis获取黑名单错误:%v", err)
		return value, false, err
	}
	value = value.ParseBlackType(res)
	return value, true, nil
}
