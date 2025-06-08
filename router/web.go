package routerg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"tgwp/configs"
	"tgwp/global"
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
		rg.POST("/test", middleware.Limiter(rate.Every(time.Hour)*1, 10), api.Template)

		rg.GET("/signin-list", middleware.Limiter(rate.Every(time.Minute)*5, 8), middleware.Authentication(global.ROLE_USER), api.SigninList)
		rg.POST("/signin", middleware.Limiter(rate.Every(time.Minute)*5, 8), middleware.Authentication(global.ROLE_USER), api.Signin)
		rg.POST("/signin-teacher", middleware.Limiter(rate.Every(time.Minute)*5, 8), middleware.Authentication(global.ROLE_USER), api.SigninTeacher)

		rg.GET("/auto-list", middleware.Limiter(rate.Every(time.Second)*2, 4), middleware.Authentication(global.ROLE_USER), api.GetAutoList)
		rg.POST("/auto-setting", middleware.Limiter(rate.Every(time.Minute)*3, 5), middleware.Authentication(global.ROLE_USER), api.AutoSetting)
	})

	// 注册文件上传相关路由组
	routeManager.RegisterFileRoutes(func(rg *gin.RouterGroup) {
		rg.POST("/upload", middleware.Limiter(rate.Every(time.Minute)*3, 5), middleware.Authentication(global.ROLE_GUEST), api.UploadFile)
	})

	// 注册登录相关路由组
	routeManager.RegisterLoginRoutes(func(rg *gin.RouterGroup) {
		rg.POST("/send-code", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.SendCode)
		rg.POST("/register", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.Register)
		rg.POST("/login", middleware.Limiter(rate.Every(time.Minute)*4, 4), api.Login)
		rg.POST("/refresh-token", middleware.Limiter(rate.Every(time.Second)*4, 8), api.RefreshToken)

		rg.GET("/test", middleware.Limiter(rate.Every(time.Second)*2, 5), middleware.Authentication(global.ROLE_SUPER_ADMIN), api.TokenTest)
	})

	// 注册用户相关路由组
	routeManager.RegisterUserRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/info", middleware.Limiter(rate.Every(time.Second)*20, 40), api.GetUserInfo)
		rg.GET("/my-info", middleware.Limiter(rate.Every(time.Second)*5, 10), middleware.Authentication(global.ROLE_GUEST), api.GetMyUserInfo)
		// 获取和修改用户资料
		rg.GET("/profile", middleware.Limiter(rate.Every(time.Second)*10, 20), api.GetProfile)
		rg.POST("/profile", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_GUEST), api.SetProfile)
		rg.POST("/role", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_GUEST), api.SetRole)
	})

	routeManager.RegisterMessageRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/count", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_GUEST), api.GetMessageCount)
		rg.GET("/list", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_GUEST), api.GetMessageList)
		rg.POST("/read", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_GUEST), api.MarkReadMessage)
	})

	// 注册帖子相关路由组
	routeManager.RegisterPostRoutes(func(rg *gin.RouterGroup) {
		rg.POST("/create", middleware.Limiter(rate.Every(time.Minute)*3, 3), middleware.Authentication(global.ROLE_USER), api.CreatePost)
		rg.POST("/edit", middleware.Limiter(rate.Every(time.Minute)*3, 3), middleware.Authentication(global.ROLE_USER), api.EditPost)
		rg.POST("/delete", middleware.Limiter(rate.Every(time.Minute)*3, 3), middleware.Authentication(global.ROLE_USER), api.DeletePost)

		rg.GET("/detail", middleware.Limiter(rate.Every(time.Second)*10, 10), middleware.Authentication(global.ROLE_GUEST), api.GetPostDetail)
		rg.GET("/detail-visitor", middleware.Limiter(rate.Every(time.Second)*10, 10), api.GetPostDetailVisitor)

		rg.GET("/like-post", middleware.Limiter(rate.Every(time.Second)*50, 100), middleware.Authentication(global.ROLE_USER), api.GetLikePost)
		rg.POST("/like-post", middleware.Limiter(rate.Every(time.Second)*5, 20), middleware.Authentication(global.ROLE_USER), api.LikePost)
		// rg.GET("/info", middleware.Limiter(rate.Every(time.Second)*20, 40), api.GetUserInfo)

		rg.POST("/comment", middleware.Limiter(rate.Every(time.Second)*1, 3), middleware.Authentication(global.ROLE_USER), api.CreateComment)
		rg.GET("/comment-more", middleware.Limiter(rate.Every(time.Second)*4, 10), api.GetMoreComments)

		rg.GET("/like-comment", middleware.Limiter(rate.Every(time.Second)*50, 100), middleware.Authentication(global.ROLE_USER), api.GetLikeComment)
		rg.POST("/like-comment", middleware.Limiter(rate.Every(time.Second)*5, 20), middleware.Authentication(global.ROLE_USER), api.LikeComment)

		rg.GET("/post-more", middleware.Limiter(rate.Every(time.Second)*4, 10), api.GetMorePosts)
		rg.GET("/post-page", middleware.Limiter(rate.Every(time.Second)*4, 10), api.GetPagePosts)

		rg.POST("/feature", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_ADMIN), api.SetPostFeature)

		rg.GET("/diary-list", middleware.Limiter(rate.Every(time.Second)*4, 10), api.GetDiaryList)

		rg.GET("/search", middleware.Limiter(rate.Every(time.Minute)*6, 10), api.SearchPosts)
	})

	// 注册比赛相关路由组
	routeManager.RegisterContestRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/list", middleware.Limiter(rate.Every(time.Second)*4, 8), api.GetContestList)

		rg.POST("/create", middleware.Limiter(rate.Every(time.Minute)*3, 3), middleware.Authentication(global.ROLE_ADMIN), api.CreateContest)
		rg.GET("/detail", middleware.Limiter(rate.Every(time.Second)*4, 8), api.GetContestDetail)

		rg.POST("/booking", middleware.Limiter(rate.Every(time.Second)*2, 4), middleware.Authentication(global.ROLE_USER), api.BookingContest)
		rg.GET("/booking", middleware.Limiter(rate.Every(time.Second)*8, 20), middleware.Authentication(global.ROLE_USER), api.IsBookingContest)

		rg.POST("/recommend", middleware.Limiter(rate.Every(time.Second)*4, 8), middleware.Authentication(global.ROLE_ADMIN), api.RecommendContest)
	})
}
