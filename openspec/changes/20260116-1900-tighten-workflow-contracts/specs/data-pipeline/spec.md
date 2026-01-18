# Spec: Data Pipeline Runtime Contracts

## ADDED Requirements

## MODIFIED Requirements

### Requirement: 运行时中间数据目录可配置
工作流各阶段（Fetch/AI/Submit）读写中间文件时 MUST 以 `paths.state_dir` 作为根目录；当未配置或为空时 MUST 默认使用 `data/`。

#### Scenario: 使用默认目录
- **Given** 未设置 `paths.state_dir`
- **When** 工作流运行并写入中间文件
- **Then** 中间文件写入到 `data/` 目录下

#### Scenario: 使用自定义目录
- **Given** `paths.state_dir="var/data"`
- **When** 工作流运行并写入中间文件
- **Then** 中间文件写入到 `var/data/` 目录下

### Requirement: 中间文件命名保持一致
工作流 MUST 使用以下固定文件名作为阶段边界（位于 state_dir 下）：
- `pending_audits.jsonl`
- `pending_audits_results.jsonl`

#### Scenario: 端到端文件产出
- **Given** 存在待审核数据集
- **When** 完整运行 Fetch 与 AI 阶段
- **Then** 生成 `pending_audits.jsonl` 与 `pending_audits_results.jsonl`

## REMOVED Requirements

