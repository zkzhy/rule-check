# 任务：优化项目结构

- [ ] 移除 `data/pending_audits_risk.jsonl` 的生成逻辑 @refactor
  - [ ] 修改 `internal/orchestrator/risk_analysis.go`，停止创建/写入该文件。
  - [ ] 确保仍正常生成 `data/pending_audits_results.jsonl`。
- [ ] 将 Taxonomy 重构为 Tool 组件 @refactor
  - [ ] 将 `internal/taxonomy` 迁移到 `internal/components/tools/taxonomy`。
  - [ ] 更新 `internal/components/tools/submit` 中的 import 引用。
  - [ ] 验证 `ATT&CK.csv` 的加载仍可用（关注路径解析）。
- [ ] 重组 Prompts 目录结构 @refactor
  - [ ] 创建 `internal/components/prompts`。
  - [ ] 将 `ai/prompts/*.json` 迁移到 `internal/components/prompts/`。
  - [ ] 更新 `internal/components/prompt/risk.go`（或等价加载器）指向新路径。
- [ ] 清理目录 @cleanup
  - [ ] 删除空的 `ai/` 目录。
  - [ ] 删除空的 `internal/taxonomy` 目录。
- [ ] 验证 @test
  - [ ] 执行 `go mod tidy`。
  - [ ] 执行 `go build ./...`，确保无编译/引用错误。
  - [ ] 执行测试/小流程，确保 prompts 加载与 taxonomy 映射可用。
