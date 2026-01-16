# Project Context

## Purpose
自动化抓取“待审核”的 HTTP 漏洞数据，通过可插拔的 AI 提供者生成蓝队视角风险评分与说明，并以 JSONL 形式落地，后续将结果回写到御衡平台。

当前代码库以纯 Go 工作流实现为主：基于 Go + Eino 的统一工作流，将 Fetch → AI Audit → Submit 编排在单一 Graph 中，并通过 data/ 目录的中间文件支持断点与排查。

## Tech Stack
- Go 1.21+
  - 模块：`module audit-workflow`（见 go.mod）
  - 编排：`github.com/cloudwego/eino v0.7.18`（Graph 组件）
  - 模型组件：`github.com/cloudwego/eino-ext/components/model/ark v0.1.62`（Doubao Ark ChatModel）
  - 内部模块：
    - internal/fetch：分页获取待审核 ID 与详情，写入 JSONL。
    - internal/llm：基于 Ark ChatModel 生成风险分，支持模板渲染与速率限制。
    - internal/submit：将评分与建议提交到御衡平台审核接口。
    - internal/eino：封装 Eino Graph 工作流编排。
    - internal/httpclient：HTTP 封装，支持超时与自签证书跳过。
    - internal/config：集中配置加载与校验，支持环境变量覆盖。
- AI 提供者
  - Doubao Ark（通过 Eino Ark ChatModel 直接调用）
  - 可扩展其他 Provider（通过 internal/llm 的接口与配置扩展）
- 数据格式与存储
  - JSONL 作为模块之间的主数据交换格式：
    - data/pending_audits.jsonl：抓取输出
    - data/pending_audits_risk.jsonl：AI 评分输出
  - 运行时中间数据统一放在 data/ 目录

## Project Conventions

### Code Style
- Go
  - 采用标准 Go module 布局：cmd/ + internal/ + data/。
  - 包命名简短小写：fetch、llm、submit、eino、httpclient、config。
  - 入口程序位于 cmd/workflow/main.go。
  - 结构体围绕领域命名，保留扩展字段（ATT&CK、标签、防护建议）。
  - 日志风格：简洁进度输出，统一前缀，如 `[Fetch]`、`[Info]`、`[Success]`。
- 配置与密钥
  - 非敏感配置：config/app.json。
  - 敏感配置：config/secrets.local.json（通过 .gitignore 排除远端提交）。
  - 环境变量覆盖：YH_CONFIG、YH_SECRETS、AI_PROVIDER、AI_PROMPT_PATH、AI_DEBUG。
- 目录与文件命名
  - 保持跨模块命名对齐：pending_audits.jsonl / pending_audits_risk.jsonl。
  - 运行时中间数据统一放在 data/ 目录。

### Architecture Patterns
- 三段式审核链路（Go 工作流）
  - Fetch：抓取“待审核”HTTP 漏洞列表与详情，支持分页。
  - LLM：根据提示词模板，对每条记录生成 1–10 的风险评分；模板支持占位符。
  - Submit：将评分与建议回写到御衡平台审核接口。
- 配置驱动
  - 外部参数（API 地址、超时、模型名称、BaseURL 等）来自 JSON 配置与环境变量。
  - 通过配置切换测试/生产，业务逻辑不绑定具体环境。
- 可插拔 AI Provider
  - 通过 internal/llm 封装模型调用，按配置选择 Provider（当前为 Ark）。
- Eino Graph 工作流
  - 使用 Eino Graph 编排 Fetch → LLM → Submit，形成线性工作流。
  - 每个节点执行完写入 data/ 中间文件，便于断点续跑与定位问题。
  - 工作流入口位于 cmd/workflow/main.go。
- 中间态 JSONL 作为边界
  - 各模块之间通过 JSONL 文件解耦，便于独立运行与调试。

### Testing Strategy
- 当前以手动集成测试为主（Go）：
  - 本地准备 config/app.json 与 config/secrets.local.json（通过 .gitignore 排除提交）。
  - 运行 `go run cmd/workflow/main.go` 验证整条 Graph 链路：抓取 → 评分 → 提交。
  - 通过环境变量启用调试：`AI_DEBUG=1` 打印首条记录的 Prompt 与模型响应片段。
- 未来单元测试规划（Go）：
  - 为 internal/fetch、internal/llm、internal/submit 增加基于接口的单元测试。
  - 针对 Eino Graph 的端到端集成测试，校验节点编排与文件边界。

### Git Workflow
- 分支策略
  - 默认假设采用 main 作为主分支，feature 分支按功能命名（如 feature/add-new-provider）。
  - Bugfix 分支建议使用 bugfix/<short-description>。
- 提交习惯
  - 提交信息简洁描述变更与影响，例如：
    - `feat: add doubao sdk provider`
    - `refactor: unify jsonl schema for risk output`
  - 功能开发应包含脚本或 Go 工作流的验证说明（在 PR 描述或变更记录中）。

## Domain Context
- 业务背景
  - 御衡平台产生大量“待审核”的 HTTP 漏洞记录，需要安全工程师人工审核。
  - 本项目的目标是自动化这部分工作，从抓取、评分到提交尽量无人值守。
- 核心实体
  - 漏洞记录：包含 name、description、req_pkg、resp_pkg 等 HTTP 请求/响应上下文字段。
  - 风险评分：1–10 的整数，数字越大风险越高，由 AI 文本输出中的首个数字解析得到。
  - 蓝队视角说明：AI 生成的安全分析文本，描述风险点、攻击链与防护建议（在扩展 Provider 时重点建设）。
- 典型流程
  - 定时或手动触发抓取 → 生成 pending_audits.jsonl。
  - 基于提示词模板生成 AI 评分 → 生成 pending_audits_risk.jsonl。
  - 将评分与建议填充到御衡的审核接口 → 更新平台中的漏洞状态。

## Important Constraints
- 安全与合规
  - API Key、账号密码等敏感信息只存放在 `config/secrets.local.json` 或环境变量中，不应提交到仓库。
  - 输出文件中不包含账号密码等敏感数据，只保留业务必需的请求上下文和评分。
  - 终端日志中避免完整打印 HTTP 请求/响应，仅打印必要摘要和进度。
- 技术约束
  - Go 版本固定为 1.21+，以保证与 Eino 依赖兼容。
  - AI Provider 默认采用官方 SDK（如豆包 Ark SDK），避免通过非官方逆向接口。
  - HTTP 客户端需支持内网自签证书场景（通过 verify_ssl=false 或等价机制控制）。
- 性能与稳定性
  - AI 调用存在速率限制，通过简单 sleep 限速与重试机制避免打爆 Provider。
  - 通过 JSONL 流式写入的方式处理大批量记录，避免一次性加载全部数据到内存。

## External Dependencies
- 御衡平台
  - HTTP API：用于拉取“待审核”漏洞与提交审核结果。
  - 通过 config/app.json 中的 base_url 与认证信息访问（敏感信息仅在本地 secrets）。
- Doubao Ark（字节豆包）
  - 当前主 AI 提供者，负责生成风险评分（仅数字）。
  - 配置来自 config/app.json 的 ai 节点；密钥通过环境变量或本地 secrets 注入。
- OpenSpec 工具链
  - 通过 openspec/ 目录管理项目能力与变更规格。
  - 在较大变更（新能力、架构调整等）前先更新 specs 与 changes，保持实现与规格同步。
