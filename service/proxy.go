package service

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"llm-proxy/config"
)

type ProxyService struct {
	config   *config.Config
	proxies  map[string]*httputil.ReverseProxy
}

func NewProxyService(cfg *config.Config) *ProxyService {
	proxies := make(map[string]*httputil.ReverseProxy)
	
	for provider, providerConfig := range cfg.Providers {
		targetURL, err := url.Parse(providerConfig.BaseURL)
		if err != nil {
			log.Printf("解析 %s URL 失败: %v", provider, err)
			continue
		}
		
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		
		// 创建局部变量避免闭包问题
		providerName := provider
		targetHost := targetURL.Host
		targetScheme := targetURL.Scheme
		
		// 配置代理以支持流式响应
		proxy.FlushInterval = -1 // 禁用缓冲，立即刷新
		
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetScheme
			req.URL.Host = targetHost
			req.Host = targetHost
			
			originalPath := req.URL.Path
			trimmedPath := strings.TrimPrefix(originalPath, fmt.Sprintf("/llm/%s", providerName))
			req.URL.Path = trimmedPath
			
			if req.URL.RawQuery != "" {
				req.URL.RawPath = trimmedPath + "?" + req.URL.RawQuery
			}
			
			log.Printf("转发请求: %s %s -> %s%s (原始Host: %s)", 
				req.Method, 
				originalPath,
				targetHost,
				req.URL.Path,
				req.Header.Get("Host"))
		}
		
		// 修改响应以支持流式传输
		proxy.ModifyResponse = func(resp *http.Response) error {
			// 记录响应信息用于调试
			log.Printf("响应 [%s] - Status: %d, Content-Type: %s, Transfer-Encoding: %s", 
				providerName, 
				resp.StatusCode,
				resp.Header.Get("Content-Type"),
				resp.Header.Get("Transfer-Encoding"))
			
			// 对于流式响应，添加必要的头部以防止缓冲
			contentType := resp.Header.Get("Content-Type")
			if strings.Contains(contentType, "text/event-stream") {
				resp.Header.Set("X-Accel-Buffering", "no") // 防止 Nginx 缓冲
			}
			
			return nil
		}
		
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("代理错误 [%s]: %v", providerName, err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(fmt.Sprintf(`{"error": "Bad Gateway", "message": "%s"}`, err.Error())))
		}
		
		proxies[provider] = proxy
	}
	
	return &ProxyService{
		config:  cfg,
		proxies: proxies,
	}
}

func (s *ProxyService) GetProxy(provider string) (*httputil.ReverseProxy, error) {
	proxy, exists := s.proxies[provider]
	if !exists {
		return nil, fmt.Errorf("不支持的 provider: %s", provider)
	}
	return proxy, nil
}