# LLM Proxy Service

一个轻量级的大语言模型 API 转发代理服务，支持 OpenAI、Anthropic Claude、Google Gemini、Grok 等主流 LLM 服务。

## 功能特性

- ✅ **完全透传**：Headers、Body、Query Params 原样转发
- ✅ **支持流式响应**：SSE 流式响应完美支持（支持 OpenAI、Anthropic 等）
- ✅ **零配置**：无需处理认证、错误等，直接透传
- ✅ **高性能**：基于 Go + Gin 实现，使用标准库 httputil.ReverseProxy
- ✅ **自动流式检测**：自动检测并正确处理流式请求

## 使用方式

### API 路径规则

```
/llm/{provider}/*

示例：
- /llm/openai/v1/chat/completions
- /llm/anthropic/v1/messages
- /llm/google/v1beta/gemini-pro:generateContent
- /llm/grok/v1/chat/completions
```

### 请求示例

```bash
# OpenAI
curl -X POST http://localhost:9000/llm/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'

# Claude (支持流式响应)
curl -X POST http://localhost:9000/llm/anthropic/v1/messages \
  -H "x-api-key: xxx" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

## 部署方式

### 本地运行

```bash
cd llm-proxy
go mod tidy
go run cmd/main.go
```

### Docker 部署

```bash
# 构建镜像
docker build -t llm-proxy:latest .

# 导出镜像
docker save -o llm-proxy.tar llm-proxy:latest

# 使用 docker-compose 启动
docker-compose -f docker-compose-llm-proxy.yml up -d
```

### 环境变量

- `PORT`: 服务端口（默认：9000）
- `OPENAI_BASE_URL`: OpenAI API 地址（默认：https://api.openai.com）
- `ANTHROPIC_BASE_URL`: Anthropic API 地址（默认：https://api.anthropic.com）
- `GOOGLE_BASE_URL`: Google API 地址（默认：https://generativelanguage.googleapis.com）
- `GROK_BASE_URL`: Grok API 地址（默认：https://api.x.ai）

## 健康检查

```bash
curl http://localhost:9000/health
```