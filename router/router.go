package router

import (
	"github.com/gin-gonic/gin"
	"llm-proxy/controller"
)

func SetupRouter(handler *controller.ProxyHandler) *gin.Engine {
	router := gin.New()
	
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "llm-proxy",
		})
	})
	
	llmGroup := router.Group("/llm")
	{
		llmGroup.Any("/*path", handler.Handle)
	}
	
	return router
}