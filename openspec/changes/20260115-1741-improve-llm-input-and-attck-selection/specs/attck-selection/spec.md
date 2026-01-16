# 规范：ATT&CK 两阶段候选选择

## ADDED Requirements
### Requirement: 基于 ATT&CK.csv 的两阶段候选注入
系统 MUST 基于 [ATT&CK.csv](file:///Users/chenshurong/Desktop/Ai-work/sss/ATT%26CK.csv) 采用两阶段方式为每条记录生成候选列表，并将候选列表分阶段注入到 LLM 提示词中。

#### Scenario: 两阶段候选集被注入提示词
- **Given** 系统可访问 ATT&CK.csv
- **When** 系统对某条记录执行 ATT&CK 选择
- **Then** 第一阶段提示词仅包含 tactic 候选列表
- **And** 第二阶段提示词仅包含第一阶段所选 tactic 下的 technique/sub 候选列表
- **And** 第二阶段候选列表规模受 Top-K 与长度预算限制

### Requirement: LLM 输出必须命中候选集
系统 MUST 校验 LLM 输出的 `tactic_name`、`technique_name`、`sub_technique_name`，并确保它们命中候选列表中的中文规范名称；不命中时必须按规则降级。

#### Scenario: technique 不命中时降级
- **Given** LLM 输出的 tactic_name 命中候选集
- **And** technique_name 未命中候选集
- **When** 系统处理 LLM 输出
- **Then** 系统保留 tactic_name
- **And** 清空 technique_name 与 sub_technique_name（或按配置的降级策略处理）

### Requirement: 第一阶段战术集合可裁剪
系统 MUST 支持将第一阶段的 tactic 候选集合按配置裁剪，以适配业务常见范围并控制提示词长度。

#### Scenario: 仅使用指定的战术集合
- **Given** 配置指定第一阶段仅包含 {侦察, 资源开发, 初始访问}
- **When** 生成第一阶段提示词
- **Then** 提示词中的 tactic 候选只包含上述集合

### Requirement: 空候选集时允许只输出 tactic
当候选集生成失败或为空时，系统 MUST 允许 LLM 仅输出 `tactic_name`，并将 `technique_name`、`sub_technique_name` 置为空字符串。

#### Scenario: 候选集为空
- **Given** 候选集生成结果为空
- **When** 系统生成提示词并调用 LLM
- **Then** 提示词要求模型至少输出 tactic_name
- **And** 输出中 technique_name 与 sub_technique_name 允许为空字符串

## MODIFIED Requirements

## REMOVED Requirements
