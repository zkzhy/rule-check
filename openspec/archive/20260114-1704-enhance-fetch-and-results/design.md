# 设计：配置化 Fetch 查询与结构化 LLM 输出

## Eino / Agent 架构视角
- 当前 Go 工程已经通过 [internal/eino/workflow.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/eino/workflow.go) 使用 `compose.NewGraph` 将 Fetch → AI → Submit 串成单一工作流，本质上是将三个“专职 Agent”线性编排：
  - Fetch 节点：负责观察环境并拉取待审核数据；
  - AI 节点：负责读取上下文、调用 LLM 并生成审核结果；
  - Submit 节点：负责将审核结果回写御衡平台。
- 与 Eino ADK 文档中的 multi-agent 示例（如 `integration-project-manager` 中的 `agents/*.go`）相比，本项目暂不引入显式的 ChatModelAgent/ Supervisor 层，而是保持简单的三节点 Graph，输入输出通过 JSONL 文件解耦。
- 本次变更在设计层面对齐 Eino ADK 的思路：
  - 将 LLM 提示词 + 结构化输出视为一个“审核 Agent”，其输入为单条漏洞记录的上下文，输出为结构化审核维度；
  - 审核 Agent 通过 `pending_audits.jsonl` 与 `pending_audits_results.jsonl` 与其他节点解耦，便于后续扩展更多 Agent 或多阶段审核；
  - 若未来需要拆分剧情 Agent、技术 Agent 等，可以在 internal 目录下新增 `internal/agents/` 包，对齐官方示例的 `agents/*.go` 分层，但不改变本变更的对外行为。

## Fetch 设计
- 配置新增项（app.json/env 可覆盖）：
  - yuheng.list_endpoint: 字符串，默认为 `/api/lines/operation`
  - yuheng.list_method: 枚举 `GET|POST`，默认为 `GET`
  - yuheng.list_page_size: 整数，默认 1000
  - yuheng.list_filters: 任意键值对对象，用于过滤（如 name、description、vul_name、review_status、attribute_classification、category_id、type、chaitin_number、author、asset_type、asset、community_tag、id、time ranges 等）
  - yuheng.list_time_fields: 对象，定义时间范围字段名映射，如 `{"start_time":"录入时间起","end_time":"录入时间止","update_start":"更新时间起","update_end":"更新时间止"}`
  - yuheng.list_send_style: 枚举 `query|json` 指定 GET 的 query 方式或 POST 的 JSON body
- 请求构造：
  - GET + query：将 `list_filters + page_no + page_size` 序列化为 query；时间范围若存在，使用后端约定键名。
  - POST + JSON body：以 `Content-Type: application/json` 发送完整过滤对象。
- 分页：
  - 基于 `page_no/page_size` 循环，终止条件：返回数量 < page_size 或空列表。
- 调试：
  - 使用 FETCH_DEBUG 打印最终 URL、方法、payload 概要（敏感信息屏蔽），便于定位后端期望格式。

## LLM 结构化输出设计
- 输出文件：`data/pending_audits_results.jsonl`
- 记录 schema（示意，最终字段以 UI 需求为准）：
  - id: 原始记录 ID
  - data: 原始字段副本 + 以下新增字段：
    - tactics: []object（所属剧情，至少包含 id 与 name）
    - techniques: []object（对应技术，与 tactics 一一对应约束，至少包含 name，可选 id）
    - tags: []string（标签，与所属剧情/对应技术一一对应的枚举值）
    - eval_description: string（评估描述，描述攻击链与蓝队分析结论）
    - devices: []string（涉及时安全设备：WAF/全流量/IPS 等）
    - suggestion: string（加固建议，仅在原记录缺失时由 LLM 生成）
    - attack_result: string（台词攻击结果，如“成功利用/无法复现/需人工确认”）
    - risk_score: number（0.0~10.0，保留整数或一位小数；若上游已存在则复用）
    - level_id: int（台词等级，与既有技术字段约定保持一致）
    - community_tags: []string（社区标签，用于标记通用漏洞分类）
    - serial_number: string（长亭编号，用于与外部系统对齐）
    - extra_fields: object（预留给未来新增的审核维度，如“数据敏感度”、“业务重要性”，键名与配置/ UI 对齐）

## 结果拼接与跳过策略
- 结果文件的生成是“拼接/补全”行为：读取输入记录（如 `pending_audits.jsonl` 或 `pending_audits_risk.jsonl`）作为基线，将 LLM 返回的字段写入到输出 JSON 中。
- 对于输入记录中已经存在且非空的字段，系统 MUST 直接复用，不需要推送给 LLM：
  - risk_score：如果已存在，则不要求 LLM 再生成风险分。
  - suggestion（加固建议）：如果已存在，则不要求 LLM 生成该字段。
- 对于缺失字段，系统 MUST 让 LLM 按模板跑完一遍并返回结构化 JSON，然后将缺失字段合并进输出记录。

## 枚举映射（所属剧情/对应技术/标签）
- 这些字段不是自由文本：必须从既定的“剧情-技术-标签”映射中选择，保证一一对应。
- 映射来源参考 [1.txt](file:///Users/chenshurong/Desktop/Ai-work/go-audit-workflow%2016.18.32/1.txt)，其中包含：
  - 战术/剧情（id 1-14，如“侦察、初始访问、执行…”）及其可选技术列表
  - 额外分类（如 id 15-22 的安全类目）
- 实现阶段建议将该映射落地为可配置数据文件（JSON/YAML 之一），并在 LLM 提示词中要求只输出该映射内的值。
- 提示词：
  - 模板要求模型返回严格 JSON（不含多余文本），字段名称与 schema 对齐；字段集合由一个“审核维度配置”驱动：
    - 剧情/技术/标签维度：tactics、techniques、tags；
    - 既有技术字段：risk_score、level_id、community_tags、serial_number；
    - 未来扩展维度：通过配置声明需要填充到 `extra_fields` 下的键名，例如 `"data_sensitivity"`、`"business_impact"` 等；
  - 支持通过 `AI_PROMPT_PATH` 指定模板；在模板中包含上下文 `name/description/req_pkg/resp_pkg`；允许未来扩展字段。
  - 当输入记录已具备某字段（如 suggestion 或 risk_score）时，不应在提示词中要求模型输出该字段。
- 解析：
  - 先尝试直接 JSON 解析；失败时采用宽松提取（如正则或代码块提取）再重试解析。
  - 对 risk_score 进行范围裁剪并保留到结果。
- 写入：
  - 每条记录写一行 JSON，包含 `generated_at` 时间戳。
  - 输出文件与原始 `pending_audits.jsonl` 解耦，便于独立验证。

## 安全与稳健性
- 不打印敏感字段（token、api_key、密码）。
- 超时、重试与速率限制维持现有策略；失败记录可标注并继续处理。
