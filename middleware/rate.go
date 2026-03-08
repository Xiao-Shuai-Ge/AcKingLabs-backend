package middleware

import (
	"fmt"
	"strings"
	"sync"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/response"
	"tgwp/utils/jwtUtils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiters sync.Map

// Limiter 限流中间件
func Limiter(r rate.Limit, b int) gin.HandlerFunc {
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

		key := fmt.Sprintf("%s|%v|%d", identity, r, b)

		zlog.CtxDebugf(ctx, "调试：生成的令牌桶键：%v", key)

		// 为每个身份和限流配置创建独立的限流器
		limiter, ok := limiters.Load(key)
		if !ok {
			limiter = rate.NewLimiter(r, b)
			limiters.Store(key, limiter)
		}

		if !limiter.(*rate.Limiter).Allow() {
			zlog.CtxInfof(ctx, "请求过于频繁!")
			response.NewResponse(c).Error(response.REQUEST_FREQUENTLY)
			c.Abort()
			return
		}
		c.Next()
	}
}
