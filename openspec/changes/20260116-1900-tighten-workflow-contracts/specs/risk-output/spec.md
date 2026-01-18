# Spec: Risk Output Score Semantics

## ADDED Requirements

### Requirement: 风险分必须可判定为“可提交”
Submit 阶段 MUST 仅在存在合法风险分时提交审核结果；当风险分缺失或非法时 MUST 跳过提交并输出可诊断日志。

#### Scenario: 风险分缺失
- **Given** results 记录中不存在 `data.risk_score`
- **When** Submit 阶段处理该记录
- **Then** 跳过该记录的提交

## MODIFIED Requirements

### Requirement: 风险分数范围为 1..10
管道中 `risk_score` MUST 为整数，取值范围为 1..10。

#### Scenario: 解析到范围外数值
- **Given** 模型输出包含 `risk_score=100`
- **When** 解析与归一化风险分
- **Then** 将其夹逼为 10

#### Scenario: 解析到非正数
- **Given** 模型输出包含 `risk_score=0`
- **When** Submit 阶段检查可提交性
- **Then** 该记录被视为非法并跳过提交

## REMOVED Requirements

