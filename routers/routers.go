package routers

import (
	"github.com/gin-gonic/gin"
	"islet/controller"
	"islet/middleware"
	"net/http"
)

func SetupRouter() *gin.Engine {
	// 使用自定义的中间件创建路由
	//gin.SetMode(gin.DebugMode)
	//r := gin.New()
	//r.Use(logger.GinLogger(), logger.GinRecovery(true), middleware.RateLimitMiddleware(2*time.Second, 1))

	// 使用默认路由
	r := gin.Default()

	r.LoadHTMLFiles("./templates/index.html")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// TODO: 对每个请求进行压力测试
	v1 := r.Group("/api/v1")
	v1.POST("/signup", controller.SignUpHandler)
	v1.POST("/login", controller.LoginHandler)

	v1.GET("/refresh_token", controller.RefreshTokenHandler)

	v1.Use(middleware.JWTAuthMiddleware())
	{
		// community related operation
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		// post related operation
		v1.POST("/post", controller.CreatePostHandler)    // create
		v1.GET("/post/:id", controller.PostDetailHandler) // read

		// 从MySQL中查询帖子列表，按照创建的时间顺序进行查询
		v1.GET("/post", controller.PostListHandler)

		// 按照分数或时间查询帖子列表（从Redis缓存中获取）
		v1.GET("/post2", controller.GetPostListHandler)

		v1.POST("/vote", controller.VoteHandler) // post vote
	}

	// 404 no route
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})

	return r
}
