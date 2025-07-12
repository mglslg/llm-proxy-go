package controller

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"llm-proxy/service"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
}

func NewProxyHandler(proxyService *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
	}
}

func (h *ProxyHandler) Handle(c *gin.Context) {
	path := c.Param("path")
	// 去除前导斜杠
	path = strings.TrimPrefix(path, "/")
	pathSegments := strings.SplitN(path, "/", 2)
	
	if len(pathSegments) < 1 || pathSegments[0] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid path",
			"message": "Provider not specified",
		})
		return
	}
	
	provider := pathSegments[0]
	
	proxy, err := h.proxyService.GetProxy(provider)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Provider not found",
			"message": err.Error(),
		})
		return
	}
	
	// 读取请求体并检查是否为流式请求
	if c.Request.Method == "POST" && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			// 重新设置请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			// 对于 Anthropic，如果请求体包含 stream: true，设置正确的 Accept 头
			if provider == "anthropic" && (bytes.Contains(bodyBytes, []byte(`"stream":true`)) || bytes.Contains(bodyBytes, []byte(`"stream": true`))) {
				c.Request.Header.Set("Accept", "text/event-stream")
			}
		}
	}
	
	proxy.ServeHTTP(c.Writer, c.Request)
}