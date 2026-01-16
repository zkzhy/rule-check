1. [ ] **Structure**: 创建目录结构 `internal/components/{model,prompt,parser,tools}`, `internal/orchestrator`, `internal/types`。
2. [ ] **Types**: 在 `internal/types/schema.go` 中定义 `RiskRecord`, `ReviewData` 等核心数据结构 (State)。
3. [ ] **Component - Model**: 迁移 `internal/llm/llm.go` 中的 `createChatModel` 到 `internal/components/model`，适配 Eino 接口。
4. [ ] **Component - Prompt**: 将 `loadPromptTemplate` 封装为 `internal/components/prompt`，实现 Eino Template 接口。
5. [ ] **Component - Tools**: 
    - 将 `internal/taxonomy` 移入 `internal/components/tools/taxonomy`。
    - 将 `internal/submit` 的逻辑封装为 `internal/components/tools/submit` (作为 Tool 或 Function)。
6. [ ] **Component - Parser**: 将 `parseStructuredJSON` 和 `applyStructuredFields` 封装为 `internal/components/parser`，实现 OutputParser 接口。
7. [ ] **Orchestrator**: 在 `internal/orchestrator/graph.go` 中构建 `RiskAnalysisGraph`，串联 Fetcher -> Analyzer(Chain) -> Submitter。
8. [ ] **Entrypoint**: 修改 `cmd/workflow/main.go`，使用 `orchestrator.NewGraph().Run(ctx)` 替代旧的 `llm.Run`。
9. [ ] **Verify**: 运行 `go run`，确保输出文件 `pending_audits_results.jsonl` 内容与重构前一致，提交逻辑正确。
10. [ ] **Cleanup**: 删除旧的 `internal/llm` 和 `internal/submit` 目录。
