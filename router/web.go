package routerg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"tgwp/configs"
	"tgwp/internal/api"
	"tgwp/log/zlog"
	"tgwp/manager"
	"tgwp/middleware"
	"time"
)

// RunServer 启动服务器 路由层
func RunServer() {
	r, err := listen()
	if err != nil {
		zlog.Errorf("Listen error: %v", err)
		panic(err.Error())
	}
	r.Run(fmt.Sprintf("%s:%d", configs.Conf.App.Host, configs.Conf.App.Port)) // 启动 Gin 服务器
}

// listen 配置 Gin 服务器
func listen() (*gin.Engine, error) {
	r := gin.Default() // 创建默认的 Gin 引擎
	// 注册全局中间件（例如获取 Trace ID）
	manager.RequestGlobalMiddleware(r)
	//配置静态路由，用于访问上传的文件
	r.Static("/uploads", "uploads")
	// 创建 RouteManager 实例
	routeManager := manager.NewRouteManager(r)
	// 注册各业务路由组的具体路由
	registerRoutes(routeManager)
	return r, nil
}

// registerRoutes 注册各业务路由的具体处理函数
func registerRoutes(routeManager *manager.RouteManager) {

	// 注册通用路由组
	routeManager.RegisterCommonRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/test", middleware.Limiter(rate.Every(time.Minute)*5, 5), api.Template)
	})

	// 注册登录相关路由组
	routeManager.RegisterLoginRoutes(func(rg *gin.RouterGroup) {
		rg.POST("/send-code", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.SendCode)
		rg.POST("/register", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.Register)
		rg.POST("/login", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.Login)
		rg.POST("/refresh-token", middleware.Limiter(rate.Every(time.Second)*4, 8), api.RefreshToken)

		rg.GET("/test", middleware.Limiter(rate.Every(time.Second)*2, 5), middleware.Authentication, api.TokenTest)
	})

	// 注册用户相关路由组
	routeManager.RegisterUserRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/info", middleware.Limiter(rate.Every(time.Second)*20, 40), api.GetUserInfo)
		rg.GET("/my-info", middleware.Limiter(rate.Every(time.Second)*5, 10), middleware.Authentication, api.GetMyUserInfo)
		// 获取和修改用户资料
		rg.GET("/profile", middleware.Limiter(rate.Every(time.Second)*10, 20), api.GetProfile)
		rg.POST("/profile", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication, api.SetProfile)
		rg.POST("/role", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication, api.SetRole)
	})

	// 注册帖子相关路由组
	routeManager.RegisterPostRoutes(func(rg *gin.RouterGroup) {
		rg.POST("/create", middleware.Limiter(rate.Every(time.Minute)*1, 300), middleware.Authentication, api.CreatePost)

		// rg.GET("/info", middleware.Limiter(rate.Every(time.Second)*20, 40), api.GetUserInfo)
	})
}
