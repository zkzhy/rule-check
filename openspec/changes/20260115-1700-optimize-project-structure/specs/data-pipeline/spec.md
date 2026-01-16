# 规范：数据流水线优化

## ADDED Requirements

## MODIFIED Requirements
### Requirement: 精简数据输出
风险分析阶段 MUST NOT 生成冗余的中间文件。

#### Scenario: 执行工作流
- **Given** 存在待审核数据集
- **When** 风险分析阶段完成
- **Then** 仅生成 `data/pending_audits_results.jsonl`
- **And** 不生成 `data/pending_audits_risk.jsonl`

## REMOVED Requirements
### Requirement: 风险中间文件
- 移除生成 `data/pending_audits_risk.jsonl` 的要求。
