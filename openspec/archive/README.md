# OpenSpec 归档

本目录用于存放已完成或不再活跃的 OpenSpec 变更记录（归档摘要）。

## 已归档变更

### 2026-01-15
- **20260114-1527-add-go-eino-workflow**：首次引入 Go Eino 工作流实现。
- **20260114-1528-migrate-to-pure-go**：迁移到纯 Go 的计划与任务拆解。
- **20260114-1704-enhance-fetch-and-results**：Fetch 配置化与结构化结果拼接增强。
- **20260115-1741-improve-llm-input-and-attck-selection**：精简 LLM 入参 + ATT&CK 两阶段选择（tactic → technique/sub）。
  - 影响范围：仅 AI 风险分析阶段；Fetch/Submit 的 API 形态不变。
  - 入参裁剪：按白名单拼装上下文（name/description/poc/req/resp），并支持总长度与单字段长度预算。
  - 两阶段选择：第一阶段只给战术候选列表；第二阶段仅注入所选战术下的 Top-K 技术/子技术候选。
  - 输出校验/降级：tactic 强制使用第一阶段结果；technique/sub 不命中候选集则清空。
  - 配置新增：`ai.context.*` 与 `ai.attck.*`（csv_path、tactic_allowlist、technique_top_k、candidate_max_runes、sub_max_per_technique）。
  - 验证：`go test ./...`、`go build ./...`、`go vet ./...`。

## 当前活跃变更
见 `openspec/changes/`。
