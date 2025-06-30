package cacheUtils

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"tgwp/global"
	"time"
)

func Get(s string) (string, error) {
	value, err := global.Rdb.Get(context.Background(), s).Result()
	//zlog.CtxDebugf(context.Background(), "Get key:%s, value:%s, err:%v", s, value, err)
	if errors.Is(err, redis.Nil) {
		return "", err
	} else if err != nil {
		return "", err
	}
	return value, nil
}

func Set(key string, value string, timeout time.Duration) error {
	return global.Rdb.Set(context.Background(), key, value, timeout).Err()
}

func Remove(key string) error {
	return global.Rdb.Del(context.Background(), key).Err()
}
