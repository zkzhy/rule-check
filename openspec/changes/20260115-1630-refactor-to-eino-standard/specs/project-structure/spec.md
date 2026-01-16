## ADDED Requirements

#### Scenario: 标准化目录结构
- **Given** 当前项目结构扁平且逻辑耦合。
- **When** 执行重构。
- **Then** 项目应包含 `internal/components` (原子能力) 和 `internal/orchestrator` (编排) 目录。
- **And** `internal/llm` 应被拆解并移除。

#### Scenario: Eino Graph 编排
- **Given** 风险分析流程。
- **When** 运行工作流。
- **Then** 应通过 Eino `Graph` 或 `Chain` 执行：`Input -> Prompt -> Model -> Parser -> Output`。
- **And** 这种编排方式应支持通过 Eino 可视化工具（如果未来引入）查看拓扑。

#### Scenario: 组件化 Prompt 与 Model
- **Given** 多种 AI Provider 需求。
- **When** 初始化 Graph。
- **Then** Model 组件应通过工厂模式按配置加载（Doubao/OpenAI）。
- **And** Prompt 模板应通过 Eino 的 Template 机制加载。

## MODIFIED Requirements

#### Scenario: 保持业务逻辑一致性
- **Given** 现有的 `taxonomy` 映射规则和 `submit` 校验逻辑。
- **When** 迁移到新架构。
- **Then** `OutputParser` 或后续节点必须严格执行相同的 ID 映射和字段校验。
- **And** 最终输出的 JSONL 文件内容与重构前一致。
