package routers

import (
	"backend/controller"
	"backend/logger"
	"backend/middlewares"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
	"sync"
	"time"
)

// SetupRouter 设置路由+
func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // 设置成发布模式
	}

	//初始化 gin Engine  新建一个没有任何默认中间件的路由
	r := gin.New()
	//设置中间件
	r.Use(
		logger.GinLogger(),
		logger.GinRecovery(true), // Recovery 中间件会 recover掉项目可能出现的panic，并使用zap记录相关日志
		//middlewares.RateLimitMiddleware(500*time.Microsecond, 50), //
		middlewares.RateLimitMiddleware(1*time.Second, 800), // 每秒钟添加一百个令牌  全局限流

	)

	r.LoadHTMLFiles("templates/index.html") // 加载html
	r.Static("/static", "./static")         // 加载静态文件
	r.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})

	// 注册swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v2")

	// 注册登陆业务
	v1.POST("/login", controller.LoginHandler)               // 登陆业务
	v1.POST("/signup", controller.SignUpHandler)             // 注册业务
	v1.GET("/refresh_token", controller.RefreshTokenHandler) // 刷新accessToken
	//go controller.ConsumeUploadTasks("upload_image")
	//go controller.ConsumeProcessTasks("process_image")

	// 使用 WaitGroup 等待任务消费者完成
	var wg sync.WaitGroup
	wg.Add(3) // 我们要等待两个消费者 goroutine

	go func() {
		defer wg.Done()
		controller.ConsumeUploadTasks("upload_image")
	}()

	go func() {
		defer wg.Done()
		controller.ConsumeProcessTasks("process_image")
	}()

	//go func() {
	//	defer wg.Done()
	//	http.ListenAndServe("127.0.0.1:8080", nil)
	//}()

	// 中间件
	v1.Use(
		middlewares.JWTAuthMiddleware(),
		//middlewares.UserRateLimitMiddleware(redis.RDB, 5, time.Minute), // 用户级限流，每分钟最多 5 次请求
	) // 应用JWT认证中间件
	{
		v1.POST("/createCategory", controller.CreateCategoryHandler)                // 创建分类
		v1.GET("/category", controller.GetCategoryListHandler)                      // 获取分类列表
		v1.GET("/categoryDetail/:category_id", controller.GetCategoryDetailHandler) // 获取分类中的作品
		v1.DELETE("/deleteCategory/:category_id", controller.DeleteCategoryHandler) // 删除分类
		v1.DELETE("/deleteWorks", controller.DeleteWorksHandler)                    // 删除作品，传入列表

		v1.POST("/uploadWork", controller.UploadWorkHandler) // 上传作品

		//go controller.ConsumeProcessTasks("process_image")
		v1.POST("/processImage", controller.ProcessImageHandler)
		// 图像处理任务
		//v1.POST("/task/submit", controller.SubmitTaskHandler) // 提交图像处理任务
		v1.GET("/task/status", controller.TaskStatusHandler) // 查询图像处理任务状态

	}

	pprof.Register(r) // 注册pprof相关路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
