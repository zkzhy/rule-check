# 规范：LLM 输入裁剪

## ADDED Requirements
### Requirement: 受控的上下文裁剪与拼装
系统 MUST 在调用 LLM 前，将单条漏洞记录构造为“精简上下文”，且不得将整条原始 JSON 直接拼入提示词。

#### Scenario: 构造精简上下文
- **Given** 输入记录包含大量字段（含请求/响应、证据、描述等）
- **When** 系统构造 LLM 上下文
- **Then** 上下文只包含白名单字段（如 name、description、PoC/证据、关键 req/resp 片段）
- **And** 对超长字段按策略截断

### Requirement: 长度预算可配置
系统 MUST 支持配置化的长度预算（单字段与总长度），以控制 token 成本并避免上下文过长导致的生成不稳定。

#### Scenario: 预算生效
- **Given** 总长度预算为 N
- **When** 输入记录的原始文本总长度超过 N
- **Then** 构造后的上下文长度不超过 N

## MODIFIED Requirements

## REMOVED Requirements

