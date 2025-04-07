package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"sync"
	"tgwp/log/zlog"
	"tgwp/response"
)

var limiters sync.Map

// Limiter 限流中间件
func Limiter(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := zlog.GetCtxFromGin(c)
		ip := c.ClientIP()

		// 生成唯一键，组合IP、速率和桶大小
		key := fmt.Sprintf("%s|%v|%d", ip, r, b)

		// 为每个IP和限流配置创建独立的限流器
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
