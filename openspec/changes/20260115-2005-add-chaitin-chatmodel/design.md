# 设计：Chaitin Provider 接入（OpenAI-Compatible）

## 设计目标
- 复用现有 OpenAI-Compatible HTTP 实现，避免引入新 SDK。
- 通过 `ai.provider=chaitin` 提供明确语义的 provider 名称，减少误配。
- 将“默认访问 Chaitin”落在默认 `config/app.json` 上，而不是硬编码到程序逻辑里。

## API 兼容性与 URL 规则
Chaitin 提供 OpenAI-Compatible 接口形态：
- 请求：`POST {base}/chat/completions`
- 鉴权：`Authorization: Bearer <api_key>`

base_url 允许两种写法：
- `https://aiapi.chaitin.net/v1`
- `https://aiapi.chaitin.net`

内部拼接规则：
- 若 base_url 以 `/v1` 结尾，则直接拼接 `/chat/completions`
- 否则拼接 `/v1/chat/completions`

## 配置与密钥管理
- 非敏感项（provider/model/base_url/timeout）放在 `config/app.json`。
- `api_key` 必须放在 `config/secrets.local.json` 的 `ai.api_key`，运行时合并进配置对象。
- 日志输出必须避免包含密钥内容。

## 向后兼容
- 保持 `doubao-ai/ark` 与 `openai-compatible/deepseek` 的行为不变。
- 新增 `chaitin` 仅为别名增强，不影响既有配置。

