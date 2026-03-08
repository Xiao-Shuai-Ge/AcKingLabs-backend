package middleware

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/response"
	"tgwp/utils/jwtUtils"
	"tgwp/utils/ratelimiter"
	"time"

	"github.com/gin-gonic/gin"
)

type limiterEntry struct {
	limiter  *ratelimiter.Limiter
	lastUsed int64
}

var limiters sync.Map

// Limiter 限流中间件
func Limiter(r ratelimiter.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := zlog.GetCtxFromGin(c)
		ip := c.ClientIP()

		userID := jwtUtils.GetUserId(c)
		if userID == "" {
			authorization := c.GetHeader("Authorization")
			if authorization != "" {
				list := strings.Split(authorization, " ")
				if len(list) == 2 {
					data, err := jwtUtils.IdentifyToken(list[1])
					if err == nil && data.Class == global.AUTH_ENUMS_ATOKEN && data.Userid != "" {
						userID = data.Userid
					}
				}
			}
		}

		identity := ""
		if userID != "" {
			identity = fmt.Sprintf("uid:%s", userID)
		} else {
			identity = fmt.Sprintf("ip:%s", ip)
		}

		routeKey := c.FullPath()
		if routeKey == "" {
			routeKey = c.Request.URL.Path
		}
		key := fmt.Sprintf("%s|%s|%s", identity, c.Request.Method, routeKey)

		// zlog.CtxDebugf(ctx, "调试：生成的令牌桶键：%v", key)

		// 为每个身份和限流配置创建独立的限流器
		entry, ok := limiters.Load(key)
		if !ok {
			entry = &limiterEntry{
				limiter:  ratelimiter.NewLimiter(r, b),
				lastUsed: nowMinutes(),
			}
			limiters.Store(key, entry)
		}

		le := entry.(*limiterEntry)
		now := nowMinutes()
		last := atomic.LoadInt64(&le.lastUsed)
		if now-last >= 30 {
			atomic.StoreInt64(&le.lastUsed, now)
		}

		if !le.limiter.Allow() {
			zlog.CtxInfof(ctx, "请求过于频繁!")
			response.NewResponse(c).Error(response.REQUEST_FREQUENTLY)
			c.Abort()
			return
		}
		c.Next()
	}
}

func CleanupLimiters() {
	for i := 0; i < 5; i++ {
		keys := sampleLimiterKeys(5)
		if len(keys) == 0 {
			return
		}

		cutoff := nowMinutes() - 60 // 超过一小时删除
		deleted := 0
		for _, key := range keys {
			if entry, ok := limiters.Load(key); ok {
				le := entry.(*limiterEntry)
				last := atomic.LoadInt64(&le.lastUsed)
				if last <= cutoff {
					limiters.Delete(key)
					deleted++
					// zlog.Debugf("删除测试：%v", key)
				}
			}
		}

		if float64(deleted)/float64(len(keys)) <= 0.4 {
			return
		}
	}
}

func sampleLimiterKeys(limit int) []string {
	if limit <= 0 {
		return nil
	}
	keys := make([]string, 0, limit)
	count := 0
	limiters.Range(func(key, value any) bool {
		k, ok := key.(string)
		if !ok {
			return true
		}
		count++
		if len(keys) < limit {
			keys = append(keys, k)
			return true
		}
		j := rand.Intn(count)
		if j < limit {
			keys[j] = k
		}
		return true
	})
	return keys
}

func nowMinutes() int64 {
	return time.Now().Unix() / 60
}
