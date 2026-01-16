# Change: Add Go Eino workflow for HTTP vulnerability audit

## Why
当前项目的 HTTP 漏洞审核链路已经通过 Python 脚本实现完整的 Fetch → AI 风险评分 → Submit 流程，但仍是三段式脚本串联，缺少统一的编排与可观测能力。在引入 Go 与 Eino 的场景下，需要将这条 LLM 审核链路迁移为基于 Graph 的工作流，以便统一控制重试、限流与中间态数据。

## What Changes
- 在项目根目录下新增 Go 工作流工程，将其作为基于 Eino 的主实现入口，使用 Graph 编排 Fetch、AI、Submit 三个节点。
- 定义与现有 Python JSONL 格式兼容的 Go 结构体，用于表示待审核记录、带风险分的记录以及提交结果，并在 Graph 节点中读写 data/pending_audits.jsonl 与 data/pending_audits_risk.jsonl。
- 在 Go 工程中集成 Eino 及必要扩展（如模型组件与工作流编排），封装 internal/fetch、internal/ai、internal/submit、internal/eino 等包以对齐 openspec/project.md 中的架构约定。
- 通过配置与环境变量（config/app.json 与 config/secrets.local.json）驱动 Go 工作流中的 API 地址、认证信息、AI Provider 以及提示词路径，保持与 Python 侧行为一致。
- 在 Go 工作流中补充基础的错误处理与日志输出，确保在 Fetch、AI 或 Submit 失败时能够中断或部分重试，并通过中间 JSONL 文件辅助排查。

## Impact
- Affected specs: vuln-audit-workflow
- Affected code:
  - 新增 Go 工作流目录（例如 go-audit-workflow/），包含 cmd/workflow/main.go 与 internal/* 包。
  - 复用并对齐现有 Python 侧的配置与 JSONL 数据格式。
  - 未来在需求扩展时，Python 与 Go 两套实现需要共同遵循本次新增的 vuln-audit-workflow 能力规范。

