package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"llm-proxy/config"
	"llm-proxy/controller"
	"llm-proxy/router"
	"llm-proxy/service"
)

func main() {
	// 设置 Gin 为 release 模式
	gin.SetMode(gin.ReleaseMode)
	
	cfg := config.LoadConfig()
	
	proxyService := service.NewProxyService(cfg)
	proxyHandler := controller.NewProxyHandler(proxyService)
	
	r := router.SetupRouter(proxyHandler)
	
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}
	
	go func() {
		log.Printf("LLM 代理服务启动在端口 %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务失败: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务关闭异常:", err)
	}
	
	log.Println("服务已关闭")
}