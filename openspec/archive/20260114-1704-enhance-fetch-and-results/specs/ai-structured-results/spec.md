## ADDED Requirements

### Requirement: Structured JSON output from LLM
系统 MUST 将每条待审核记录的蓝队视角字段以结构化 JSON 形式输出到 `data/pending_audits_results.jsonl`，并与原始记录解耦，且输出 schema MUST 支持在不破坏已有字段（剧情/技术/标签、risk_score、suggestion 等）的前提下，为未来审核维度预留扩展空间（例如通过 extra_fields 对象承载新字段）。

#### Scenario: Strict JSON prompt and parsing
- 给定提示词模板要求模型只返回 JSON（字段包含 tactics、techniques、tags、eval_description、devices、suggestion、attack_result、risk_score、level_id、community_tags、serial_number）
- 当生成模型输出
- 则系统能够解析 JSON 并按每条记录写入一行结果

#### Scenario: Fallback on non-JSON responses
- 给定模型在 JSON 前后返回了额外文本
- 当首次解析失败
- 则系统应提取 JSON 片段并重试解析
- 且解析成功后继续写入输出

#### Scenario: Risk score clamping and typing
- 给定 risk_score 超出 0..10
- 当写入输出
- 则 risk_score 会被裁剪到 0..10 并以 number 序列化

#### Scenario: Reuse existing risk_score when present
- 给定输入记录已经包含上游阶段的 risk_score
- 当生成结果输出
- 则系统直接复用该 risk_score，不要求 LLM 再生成该字段

#### Scenario: Configurable prompt path and evolvable fields
- 给定 AI_PROMPT_PATH 指向新的模板
- 且审核维度配置中新增一个字段（例如 data_sensitivity）映射到 extra_fields.data_sensitivity
- 当运行 LLM
- 则系统使用新模板并按模板定义产出字段
- 且对于新增字段，输出记录中会在 data.extra_fields 下出现对应键名；缺失字段使用空值或合理默认值

### Requirement: Do not request suggestion when already present
系统 MUST 在输入记录已存在 suggestion（加固建议）时，跳过该字段的 LLM 生成与推送。

#### Scenario: Reuse existing suggestion
- 给定输入记录 suggestion 非空
- 当构建提示词并调用 LLM
- 则提示词不要求模型输出 suggestion
- 且输出结果中的 suggestion 与输入保持一致

### Requirement: Taxonomy-constrained tactics, techniques, and tags
系统 MUST 将“所属剧情/对应技术/标签”限制为预定义映射中的值，且三者保持一一对应关系。

#### Scenario: Select values from predefined mapping
- 给定从 [1.txt](file:///Users/chenshurong/Desktop/Ai-work/go-audit-workflow%2016.18.32/1.txt) 导出的枚举映射
- 当产出 results 字段 tactics、techniques、tags
- 则每个输出值都必须存在于映射中

#### Scenario: Enforce one-to-one correspondence
- 给定某个 technique 在映射中只属于唯一的 tactic
- 当模型输出该 technique
- 则输出必须包含匹配的 tactic 以及对应的 tag
