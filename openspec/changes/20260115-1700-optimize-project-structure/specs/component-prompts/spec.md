# 规范：Prompts 组件资源

## ADDED Requirements
### Requirement: 内部 Prompts 组件
提示词资源 MUST 位于 `internal/components` 结构内，以保证更好的封装性与可扩展性。

#### Scenario: 提示词位置
- **Given** 项目文件系统结构已调整
- **When** 查找提示词模板
- **Then** 在 `internal/components/prompts`（或其子目录）下找到

## MODIFIED Requirements
### Requirement: 提示词加载
应用配置与加载逻辑 MUST 引用新的提示词位置。

#### Scenario: 加载提示词
- **Given** 应用启动
- **When** 提示词组件初始化
- **Then** 能从 `internal/components/prompts` 正确读取模板

## REMOVED Requirements
### Requirement: 根目录 AI 目录
- 移除位于根目录的 `ai/prompts` 目录。
