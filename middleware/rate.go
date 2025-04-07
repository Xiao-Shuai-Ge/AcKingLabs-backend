package middleware

import (
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
		ip := c.ClientIP() // 获取客户端 IP

		// 为每个 IP 创建独立的限流器
		limiter, ok := limiters.Load(ip)
		if !ok {
			limiter = rate.NewLimiter(r, b)
			limiters.Store(ip, limiter)
		}

		zlog.CtxInfof(ctx, "ip:%s, rate:%d, burst:%d", ip, r, b)
		// 检查是否允许请求
		if !limiter.(*rate.Limiter).Allow() {
			zlog.CtxInfof(ctx, "请求过于频繁!")
			response.NewResponse(c).Error(response.REQUEST_FREQUENTLY)
			c.Abort()
			return
		}
		c.Next()
	}
}
