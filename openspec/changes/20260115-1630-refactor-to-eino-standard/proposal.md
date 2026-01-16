---
change-id: 20260115-1630-refactor-to-eino-standard
summary: 按照 Eino 官方示例重构项目结构，采用 Graph 编排与组件化设计，实现标准的 AI Agent 架构。
status: proposal
---

## 问题
当前项目结构较为扁平，`internal/llm` 承担了过多的业务逻辑（数据加载、Prompt 渲染、LLM 调用、结果解析、文件写入），不符合 Eino 框架提倡的 **Graph（编排）** + **Component（组件）** 分层设计。这导致：
1.  **耦合度高**：难以单独替换 Prompt 模板或 LLM 模型。
2.  **扩展性差**：新增“搜索”或“代码执行”节点时，需要在 `Run` 函数中堆砌逻辑。
3.  **不规范**：未利用 Eino 的 `Compose`、`Chain` 或 `Graph` 能力，仅仅是把 Eino 当作 SDK 调用。

## 目标
参考 [Eino Examples](https://github.com/cloudwego/eino-examples/tree/main/adk/multiagent/integration-project-manager)，重构为标准的 Eino 应用结构：
1.  **目录重组**：按功能模块划分 `internal/components`（原子组件）、`internal/orchestrator`（编排逻辑）、`internal/types`（领域模型）。
2.  **组件化**：
    - 将 Prompt 渲染封装为 `PromptTemplate` 组件。
    - 将 LLM 调用封装为 `ChatModel` 组件（保留多 Provider 支持）。
    - 将结果解析封装为 `OutputParser` 组件。
3.  **编排化**：使用 `eino.Graph` 串联流程：`Fetcher -> RiskAnalyzer -> ResultWriter -> Submitter`。

## 范围
- **重构**：`internal/llm`、`internal/submit`、`cmd/workflow`。
- **新增**：`internal/orchestrator`、`internal/components`、`internal/types`。
- **保持**：`internal/fetch`（作为 Input Node）、`internal/taxonomy`（作为 Tool/Resource）。

## 方案

### 1. 目录结构优化 (Standard Eino Structure)
```text
internal/
  components/       # 原子组件 (Atomic Components)
    prompt/         # 提示词模板 (实现 Eino Template 接口)
    model/          # 模型工厂 (ChatModel: Doubao/OpenAI)
    parser/         # 结果解析器 (OutputParser: String -> RiskResult)
    tools/          # 工具集 (Taxonomy Lookup, Submit Tool)
  orchestrator/     # 编排逻辑 (Orchestration)
    graph.go        # 定义 Graph 拓扑 (Fetch -> Analyze -> Submit)
  types/            # 领域模型 (Domain Models)
    schema.go       # 定义 State/Context (RiskRecord, ReviewData)
```

### 2. 核心重构点

#### A. 组件化 (Components)
- **Model**: 将 `createChatModel` 移入 `components/model`，实现工厂模式。
- **Prompt**: 将 `ai/prompts/risk.json` 加载逻辑移入 `components/prompt`。
- **Parser**: 将 JSON 解析与 Taxonomy 映射逻辑封装进 `components/parser`。

#### B. 编排 (Orchestration)
构建一个 **RiskAnalysisGraph**：
1.  **Fetcher Node**: 读取 `pending_audits.jsonl`，输出 `[]RiskRecord`。
2.  **Analyzer Chain**: 核心分析逻辑，内部为 `Prompt -> ChatModel -> Parser` 的 Chain。
3.  **Writer Node**: 将分析结果写入 `pending_audits_results.jsonl`。
4.  **Submitter Node**: 调用 Yuheng API 提交结果（可封装为 Tool，但在 Batch 模式下作为 Node 更高效）。

#### C. Tool 化思考
- **Fetch**: 保持为 Source Node (Graph 起点)，不作为 Tool，因为它负责产生数据流。
- **Submit**: 建议封装为 `SubmitTool`。在当前 Batch 场景下，它作为 Graph 的 Sink Node (终点) 运行。如果未来升级为 ReAct Agent，它可以直接被 Agent 调用。

### 3. 验证
- 保持 `go run cmd/workflow/main.go` 入口不变，但内部调用 `orchestrator.NewGraph().Run()`。
- 输出文件内容与提交行为需与重构前完全一致。

## 风险与缓解
- **风险**: 重构幅度大，可能丢失 `submit.go` 中的复杂字段映射（如 Taxonomy ID 查找、Devices 校验）。
- **缓解**: 在 `components/parser` 或 `components/tools` 中完整保留这些逻辑，并编写单元测试确保行为一致。
