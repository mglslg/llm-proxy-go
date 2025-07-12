# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概览

这是一个 **LLM API 转发代理服务**，使用 Go + Gin 构建，通过标准库 `httputil.ReverseProxy` 实现对多个 LLM Provider 的透明代理转发。

## 核心架构

### 请求流程
```
Client → [/llm/{provider}/*] → Gin Router → ProxyHandler → ProxyService → ReverseProxy → LLM Provider
```

### 路径映射规则
- `/llm/openai/*` → `https://api.openai.com/*`
- `/llm/anthropic/*` → `https://api.anthropic.com/*`
- `/llm/google/*` → `https://generativelanguage.googleapis.com/*`
- `/llm/grok/*` → `https://api.x.ai/*`

### 关键设计
1. **完全透传**：Headers、Body、Query Params 原样转发，不做任何修改
2. **无状态设计**：服务本身不存储任何状态，可水平扩展
3. **流式响应支持**：天然支持 SSE/流式传输
4. **优雅关闭**：支持信号处理和 5 秒超时

## 开发指令

### 构建和运行
```bash
# 本地开发
go mod tidy
go run cmd/main.go

# 构建二进制
go build -o llm-proxy ./cmd/main.go

# Docker 构建
docker build -t llm-proxy:latest .
docker save -o llm-proxy.tar llm-proxy:latest

# 使用 docker-compose 运行
docker-compose -f docker-compose-llm-forward.yml up -d
```

### 测试服务
```bash
# 健康检查
curl http://localhost:9000/health

# 测试 OpenAI 代理
curl -X POST http://localhost:9000/llm/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "test"}]}'
```

### 代码规范
1. **错误处理**：代理错误统一返回 502 状态码
2. **日志格式**：使用 `log.Printf` 记录关键信息
3. **配置管理**：所有配置项在 `config/config.go` 中集中管理
4. **路由注册**：新增 provider 需在 `router/router.go` 中添加路由

## 常见开发任务

### 添加新的 LLM Provider
1. 在 `config/config.go` 中添加新 Provider 的 BaseURL 配置
2. 在 `service/proxy.go` 的 `GetTargetURL` 方法中添加 case 分支
3. 在 `router/router.go` 中注册新路由：`r.Any("/llm/newprovider/*path", handler.ProxyHandler)`
4. 更新 `docker-compose-llm-forward.yml` 添加环境变量

### 修改端口
1. 修改 `docker-compose-llm-forward.yml` 中的 `PORT` 环境变量和端口映射
2. 修改 `Dockerfile` 中的 `EXPOSE` 指令

## 注意事项

### 已知问题
1. **配置不一致**：`config.go` 中 Anthropic 的 BaseURL 硬编码为 `https://api.anthropic.com`，未读取 `ANTHROPIC_BASE_URL` 环境变量
2. **缺少测试**：项目中没有单元测试或集成测试
3. **错误细分不足**：所有代理错误都返回 502，建议根据上游错误返回更精确的状态码

### 流式响应支持（2025-01-12 修复）
#### 问题描述
Anthropic Claude API 的流式响应（`"stream": true`）通过代理后无法正常工作，表现为：
- Postman 显示原始 SSE 格式而非解析后的事件流
- 响应被缓冲而非实时传输

#### 解决方案
1. **自动检测流式请求**：解析请求体，当包含 `"stream": true` 时自动为 Anthropic 请求添加 `Accept: text/event-stream` 头
2. **禁用响应缓冲**：设置 `proxy.FlushInterval = -1` 确保数据立即传输
3. **保持原始响应头**：不修改上游服务的 Content-Type，让客户端正确解析 SSE 格式
4. **添加防缓冲头**：设置 `X-Accel-Buffering: no` 防止 Nginx 等反向代理缓冲

#### 实现细节
- 在 `controller/proxy_handler.go` 中检测流式请求并自动设置必要的请求头
- 在 `service/proxy.go` 中配置 ReverseProxy 的 FlushInterval 和 ModifyResponse
- 确保响应头原样传递，仅添加必要的防缓冲设置

### 安全考虑
- 服务不处理认证，依赖客户端传递正确的 API Key
- 无访问控制或速率限制功能
- 生产环境建议前置 API Gateway 或反向代理

### 部署注意
- 使用外部 Docker 网络 `api-proxy_proxy_net`，需确保该网络存在
- 镜像基于 Alpine Linux，体积小但可能缺少某些调试工具
- 日志使用 json-file driver，注意配置 max-size 防止磁盘占满