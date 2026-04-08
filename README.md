# NovelAI Local Gateway

轻量的 Gin 后端，封装 NovelAI 文本、图像、Stories 与 OpenAI 兼容接口，方便本地工具（如 SillyTavern）直接调用。

## 功能范围

- 文本：`/api/text/*` 与 `/v1/completions`、`/v1/chat/completions`、`/v1/responses`
- 图像：`/api/image/generate`、`/api/image/director-tools`、`/api/image/encode-vibe`
- OpenAI 图像：`/v1/images/generations`（`b64_json`）
- Stories：`/api/stories/keystore`、`/api/stories/objects/*`、`/api/account/subscription`
- 模型：`/v1/models`、`/api/text/models/probe`

## 快速启动

1) 准备配置（默认读取 `config.json`）

```json
{
  "http_addr": ":18787",
  "session_ttl": "24h",
  "session_secret": "dev-secret-change-me",
  "novelai_api_base": "https://api.novelai.net",
  "novelai_image_base": "https://image.novelai.net",
  "novelai_text_base": "https://text.novelai.net",
  "log_level": "info"
}
```

2) 运行服务

```bash
go run ./cmd/server
```

3) 健康检查

```bash
curl http://127.0.0.1:18787/healthz
```

## 鉴权

受保护接口支持以下方式：

- `Authorization: Bearer pst_xxx`
- `X-API-Token: pst_xxx`
- `Authorization: Bearer sess_xxx`（会话模式）

## 常用请求示例

### Chat Completions

```bash
curl -X POST http://127.0.0.1:18787/v1/chat/completions \
  -H "Authorization: Bearer pst_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model":"xialong-v1",
    "messages":[{"role":"user","content":"Say one short word."}],
    "max_tokens":64
  }'
```

### Image Generate

```bash
curl -X POST http://127.0.0.1:18787/api/image/generate \
  -H "Authorization: Bearer pst_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt":"1girl, masterpiece",
    "model":"nai-diffusion-4-5-curated",
    "action":"img2img",
    "width":1024,
    "height":1024,
    "n_samples":1,
    "strength":0.7,
    "noise":0,
    "image":"image",
    "files":{"image":"<base64_png>"}
  }'
```

说明：
- 已支持官方图像参数顶层写法（如 `strength`、`noise`、`cfg_rescale`、`reference_*`、`director_reference_*` 等）。
- `parameters` 与 `raw_request` 仍可透传；优先级：`raw_request` > `parameters` > 顶层快捷字段。

## 构建

```bash
go build -o server.exe ./cmd/server
```

## 项目结构

- `cmd/server`：程序入口
- `internal/http`：路由与中间件
- `internal/http/handler`：HTTP Handler
- `internal/service`：业务层
- `internal/novelai`：上游 NovelAI 客户端
- `internal/model`：请求/响应模型