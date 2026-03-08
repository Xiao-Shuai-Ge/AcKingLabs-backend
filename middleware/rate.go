package middleware

import (
	"fmt"
	"strings"
	"sync"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/response"
	"tgwp/utils/jwtUtils"
	"tgwp/utils/ratelimiter"

	"github.com/gin-gonic/gin"
)

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

		zlog.CtxDebugf(ctx, "调试：生成的令牌桶键：%v", key)

		// 为每个身份和限流配置创建独立的限流器
		limiter, ok := limiters.Load(key)
		if !ok {
			limiter = ratelimiter.NewLimiter(r, b)
			limiters.Store(key, limiter)
		}

		if !limiter.(*ratelimiter.Limiter).Allow() {
			zlog.CtxInfof(ctx, "请求过于频繁!")
			response.NewResponse(c).Error(response.REQUEST_FREQUENTLY)
			c.Abort()
			return
		}
		c.Next()
	}
}
