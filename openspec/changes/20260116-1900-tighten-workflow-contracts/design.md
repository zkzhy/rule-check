# Design: Tighten Workflow Contracts

## Overview
本变更聚焦“契约收敛”，即让配置、路径、数据文件与评分语义在代码与规范中保持一致，并通过小范围重构去除重复实现。

## Decisions

### 1) Runtime Path Resolution
- 默认仍使用 `data/`，但所有阶段统一以 `paths.state_dir` 作为运行时根目录（可配置）。
- 输出文件名以现状为准：`pending_audits.jsonl` 与 `pending_audits_results.jsonl`。

### 2) Risk Score Semantics
- 风险分数定义为整数 1..10。
- AI 解析阶段允许“尽力解析”；但 Submit 阶段必须能区分“缺失/非法”与“合法”，避免把非法值静默变成 0 并提交。

### 3) Secrets Error Handling
- secrets 文件缺失：允许（视为无 secrets）。
- secrets 文件存在但不可读/JSON 非法：失败退出，让配置问题尽早暴露。

## Validation Strategy
- 单测覆盖：风险分归一化、state_dir 路径拼装、secrets 读取策略。
- 运行验证：`go test ./...`；必要时用本地 config 跑一遍 `go run cmd/workflow/main.go -mode ai` 进行最小链路验证。

