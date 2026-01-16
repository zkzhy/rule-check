## Context
项目当前通过 Python 的 1-fetch/main.py、ai/main.py 与 2-submit/main.py 实现了完整的 HTTP 漏洞自动审核链路，使用 JSONL 文件作为模块间边界。随着 Go 与 Eino 的引入，希望将这条链路迁移为单一 Graph 工作流，以提升可编排性、可观测性并为后续扩展多模型、多 Provider 与复杂分支逻辑打基础。

## Goals / Non-Goals
- Goals:
  - 使用 Go 与 Eino 构建线性 Graph，将 Fetch → AI → Submit 三个阶段编排在同一工作流中。
  - 复用现有 JSONL 数据结构与配置方式，使 Python 与 Go 实现可以并行存在并互相验证。
  - 为未来扩展多节点（如第二阶段威胁情报补充、标签预测等）保留 Graph 扩展空间。
- Non-Goals:
  - 不在本次变更中重写 Python 实现或移除现有脚本。
  - 不实现复杂的多分支、多工具调用，仅聚焦顺序三节点工作流。
  - 不在本次变更中引入额外的持久化层或队列系统。

## Decisions
- Decision: 以 Go module 形式在项目根目录下新增独立工作流工程（例如 go-audit-workflow），内部使用 cmd/workflow + internal/* 的标准布局。
  - Rationale: 与 openspec/project.md 中对 Go 结构的约定保持一致，便于后续扩展与维护。
- Decision: 使用 Eino 的 Graph/Workflow 编排 Fetch、AI、Submit 三个节点，输入输出通过本地文件系统中的 JSONL 文件完成解耦。
  - Rationale: JSONL 已在 Python 实现中验证可行，同时便于定位问题与支持断点续跑。
- Decision: 在 internal/ai 中通过 Eino 的模型组件封装对 Doubao 等 Provider 的调用，保持 Provider 可插拔能力。
  - Rationale: 与 Python 侧的 ai_providers 抽象保持语义一致，未来可以在 Go 侧扩展更多 Provider。

## Risks / Trade-offs
- Risk: Go 工作流与 Python 实现并行存在可能导致行为不一致。
  - Mitigation: 在验证阶段对比 JSONL 输出与提交流程，确保关键字段与数量一致，并通过 openspec specs 对行为进行约束。
- Risk: Eino 版本升级或接口变更对工作流造成影响。
  - Mitigation: 固定 Eino 依赖版本，并在未来升级时新增独立的 OpenSpec 变更。
- Trade-off: 继续使用文件系统 JSONL 作为中间态，而不是引入消息队列或数据库。
  - Rationale: 现有规模下 JSONL 已满足需求，引入额外基础设施会增加复杂度与运维成本。

## Migration Plan
1. 搭建 Go module 与基础目录结构，确保可以编译空 Graph。
2. 逐步将 Fetch、AI、Submit 功能以节点形式迁入 Go 工作流，保持与 Python 参数与数据格式一致。
3. 在本地环境中并行运行 Python 与 Go 实现，对比 data/pending_audits.jsonl 与 data/pending_audits_risk.jsonl 的内容。
4. 在小规模真实数据上验证 Go 工作流提交流程，确保不会破坏现有审核结果。
5. 在验证完成后，将 Go 工作流作为推荐路径记录在项目使用文档中，同时保留 Python 实现作为回退方案。

## Open Questions
- 是否需要在工作流中引入更加细粒度的错误分类与重试策略（例如针对网络错误与业务错误分别处理）。
- 是否需要在后续阶段支持多模型 Ensemble 或 A/B 实验，并在 Graph 层引入分支节点与合并节点。
- 是否需要为 Go 工作流增加统一的 metrics 与 tracing 输出，便于接入监控系统。

