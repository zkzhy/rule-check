# 规范：Taxonomy 工具组件

## ADDED Requirements
### Requirement: Taxonomy Tool 组件
ATT&CK taxonomy 逻辑 MUST 封装为与 Eino 兼容的 Tool 组件。

#### Scenario: 组件结构
- **Given** 项目目录结构已重构
- **When** 查看 `internal/components/tools`
- **Then** 存在 `taxonomy` 包
- **And** 该包提供 `LookupIDs` 能力

## MODIFIED Requirements
### Requirement: Submit 组件依赖
Submit 组件 MUST 从新的组件路径引用 taxonomy 逻辑。

#### Scenario: 构建项目
- **Given** 目录结构已重构
- **When** 编译 `internal/components/tools/submit`
- **Then** 能成功 import `internal/components/tools/taxonomy`

## REMOVED Requirements
### Requirement: 旧 taxonomy 包
- 移除位于 `internal/taxonomy` 的旧包路径。
