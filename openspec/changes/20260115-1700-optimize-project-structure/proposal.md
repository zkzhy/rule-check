# 优化项目结构与 Eino 组件组织

## 摘要
重构项目结构，使其更贴合 Eino 的组件化架构（Skills/Tools），移除冗余中间产物，并提升提示词资源的组织方式与可扩展性。

## 背景
当前代码库存在以下问题：重复生成中间文件（`data/pending_audits_risk.jsonl`）、ATT&CK taxonomy 包位置不合理（`internal/taxonomy`）、提示词目录（`ai/prompts`）缺少层级规划并不利于扩展。需要基于 Eino 的组织方式（Skills/Tools）对结构进行收敛与梳理，以提升模块化与可维护性。

## 目标
1. **移除冗余输出**：停止生成 `data/pending_audits_risk.jsonl`，避免与 results 文件重复。
2. **重构 Taxonomy**：将 `internal/taxonomy` 调整为 `internal/components/tools/taxonomy`，符合 Eino Tool/Skill 的定位。
3. **重组 Prompts**：将 `ai/prompts` 迁移到更可扩展的目录（如 `internal/components/prompts`），并更新加载逻辑。
4. **对齐 Eino 规范**：确保新的目录与依赖关系更清晰，便于后续新增工具/能力。

## 非目标
- 不改变风险分析与提交的核心业务逻辑（除路径/引用调整外）。
- 不引入新的 AI 模型或外部 Provider。

## 方案
1. **调整风险分析输出**：修改 `internal/orchestrator/risk_analysis.go`，停止写入 `data/pending_audits_risk.jsonl`。
2. **迁移 Taxonomy**：将 `internal/taxonomy` 重构到 `internal/components/tools/taxonomy`，并更新引用（主要在 `submit`）。
3. **迁移 Prompts**：将 `ai/prompts` 迁移到 `internal/components/prompts`，并更新配置/加载逻辑指向新位置。
4. **验证**：确保工作流编译、测试与运行行为保持正确。
