# 规范：Chaitin ChatModel Provider

## ADDED Requirements

### Requirement: 支持 chaitin provider
系统 MUST 支持通过 `ai.provider="chaitin"` 启用 Chaitin 的 OpenAI-Compatible ChatModel 调用方式。

#### Scenario: Provider 选择命中 chaitin
- **Given** 配置中 `ai.provider` 为 `chaitin`
- **When** 系统初始化 ChatModel
- **Then** 系统使用 OpenAI-Compatible HTTP 调用方式

### Requirement: 使用 OpenAI-Compatible 接口调用
系统 MUST 以 OpenAI-Compatible 方式向 `{base_url}/chat/completions` 发送请求，并使用 `ai.model` 作为模型名称。

#### Scenario: base_url 以 /v1 结尾
- **Given** `ai.base_url` 为 `https://aiapi.chaitin.net/v1`
- **When** 系统发起模型请求
- **Then** 实际请求路径为 `/v1/chat/completions`

#### Scenario: base_url 不含 /v1
- **Given** `ai.base_url` 为 `https://aiapi.chaitin.net`
- **When** 系统发起模型请求
- **Then** 实际请求路径为 `/v1/chat/completions`

### Requirement: 密钥只从 secrets 加载
系统 MUST 仅从 secrets 配置中的 `ai.api_key` 读取密钥，并 MUST NOT 在日志或错误信息中输出密钥明文。

#### Scenario: 运行时合并密钥
- **Given** `config/secrets.local.json` 中包含 `ai.api_key`
- **When** 系统加载配置并初始化 ChatModel
- **Then** ChatModel 使用该密钥作为 Bearer Token
- **And** 运行日志不包含密钥内容

### Requirement: 默认配置指向 Chaitin
系统 MUST 在默认 `config/app.json` 中将 AI Provider 预设为 Chaitin，以实现开箱即用的默认访问。

#### Scenario: 使用默认配置运行
- **Given** 使用默认 `config/app.json`
- **When** 系统初始化 ChatModel
- **Then** 默认 Provider 为 `chaitin`
- **And** 默认 Model 为 `gpt-5.1`

## MODIFIED Requirements

## REMOVED Requirements

