# 提案：新增 Chaitin OpenAI-Compatible ChatModel 并设为默认

## 背景
当前项目的模型组件已支持 Ark（doubao）与 OpenAI-Compatible 两类调用方式，但缺少对 `aiapi.chaitin.net` 的一键配置与“默认使用该 Provider”的约定，导致接入成本较高、容易配置错误。

用户希望：
- 使用 API 地址 `https://aiapi.chaitin.net/v1`
- 使用模型 `gpt-5.1`
- 新增一个 ChatModel Provider，并默认访问该 Provider
- 明确配置文件位置与修改方式（密钥需走 secrets）

## 目标
- 系统 SHALL 支持 `ai.provider="chaitin"`，以 OpenAI-Compatible 方式调用 Chaitin API。
- 系统 SHALL 支持在默认配置中使用 Chaitin Provider（即开箱即用默认指向 Chaitin）。
- 系统 SHALL 保持密钥只从 secrets 加载，且不在日志/输出中泄漏。
- 系统 SHALL 保持对现有 Provider（doubao/openai-compatible/deepseek）的兼容。

## 非目标
- 不调整风控/提示词/解析逻辑等与模型无关的业务行为。
- 不引入新的第三方 SDK 依赖（沿用现有 OpenAI-Compatible HTTP 实现）。
- 不在仓库中提交任何真实密钥。

## 方案概述
1) Provider 增强
- 在模型工厂中新增 `chaitin` provider 别名，将其映射到现有 OpenAI-Compatible ChatModel 实现。
- 约定 `ai.base_url` 支持带或不带 `/v1`，内部统一拼接到 `/chat/completions`。

2) 默认配置调整
- 更新 `config/app.json` 的默认 `ai.provider/ai.base_url/ai.model`，使默认配置指向 Chaitin。
- `ai.api_key` 继续放在 `config/secrets.local.json` 的 `ai.api_key`，由加载器合并到运行时配置。

3) 配置说明（对用户可见）
- 非敏感配置：`config/app.json`
- 敏感配置：`config/secrets.local.json`（应保持在 .gitignore 中）
- 环境变量覆盖：`YH_CONFIG`、`YH_SECRETS`、`AI_PROVIDER`、`AI_PROMPT_PATH` 等沿用现有约定

## 验证计划
- 单测：覆盖 Provider 选择与 URL 拼接规则（base_url 是否带 `/v1`）。
- 工具链：运行 `go test ./...`、`go build ./...`、`go vet ./...`。

