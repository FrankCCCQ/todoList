package routers

import (
	"api-gateway/internal/handler"
	"api-gateway/middlerware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewRouter(service ...interface{}) *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.Use(middlerware.InitMiddleware(service), middlerware.Cors())
	v1 := ginRouter.Group("/api/v1")
	{
		v1.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, "success")
		})
		v1.POST("/user/register", handler.UserRegister)
		v1.POST("/user/login", handler.UserLogin)

		authed := v1.Group("/")
		authed.Use(middlerware.JWT())
		{
			authed.GET("task", handler.GetTaskList)
			authed.POST("task", handler.CreateTask)
			authed.PUT("task", handler.UpdateTask)
			authed.DELETE("task", handler.DeleteTask)

		}
	}
	return ginRouter
}
