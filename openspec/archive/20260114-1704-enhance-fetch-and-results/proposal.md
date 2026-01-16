# 提案：配置化 Fetch 查询与结构化结果拼接

## 摘要
本提案引入两项能力：
- 让 fetch 模块的查询条件完全可配置（不再硬编码），并修复当前查询参数未生效导致结果固定的问题。
- 让 LLM 将提示词跑完一遍后，把返回的结构化字段按“缺失补全”的方式拼接进新的 `data/pending_audits_results.jsonl`。

## 用户需求
1) 严格按照 OpenSpec 流程执行，途中需要查询的字段全部加入并可查询；不能硬编码。
   现有 fetch 一直返回固定结果，我修改参数也不变。
2) 参考图二三，AI 提示词要完成后面字段的拼写与拼接，只需生成额外的 `pending_audits_results.jsonl`。
3) 加固建议 suggestion 字段只写入建议本身，不写“AI 自动评分”；如果原本已有 suggestion，则不推送该字段给 LLM。

## 目标
- 查询条件通过配置和环境变量注入，支持 UI 上所有筛选项及未来扩展。
- 将查询参数正确发送到后端（GET 的 query 或 POST 的 JSON body），结果随条件变化。
- LLM 输出按约定 schema 生成结构化 JSON，并以“缺失补全”方式写入 `pending_audits_results.jsonl`。

## 非目标
- 该阶段不直接改代码实现，只交付规格、设计与任务；实现将在 apply 阶段进行。

## 现状与问题
- 现状：在 [fetch.go](file:///Users/chenshurong/Desktop/Ai-work/go-audit-workflow%2016.18.32/internal/fetch/fetch.go#L129-L156) 中构造了 `params` 与 `q`，但未用于请求（GET 无 query、POST 无 body）。这会导致“参数修改无效、结果固定”。
- LLM 现状：在 [llm.go](file:///Users/chenshurong/Desktop/Ai-work/go-audit-workflow%2016.18.32/internal/llm/llm.go#L92-L179) 只提取数字评分，未输出蓝队所需的结构化字段。

## 方案概述
- Fetch：新增配置驱动的“请求方法、路径、参数映射、时间范围、分页策略”，并根据配置序列化为 GET query 或 POST JSON body，提供调试日志。
- LLM：提示词要求返回 JSON；解析失败有回退与重试策略；输出到 `data/pending_audits_results.jsonl`。
- 拼接策略：输出是对输入记录的“缺失补全”，已存在字段（如 risk_score、suggestion）直接复用，不要求 LLM 再生成。

## 验证计划
- 使用 `openspec validate 20260114-1704-enhance-fetch-and-results --strict` 验证规格结构。
- 手动核查：
  - 修改 `list_filters` 并查看请求是否包含对应参数；确认不再出现“结果固定”。
  - 对于原本已有 suggestion 的记录：确认不会把 suggestion 推给 LLM，输出仍保留原值。
  - 运行工作流后在 `data/pending_audits_results.jsonl` 中逐项确认字段补全、映射约束与 risk_score 范围裁剪。
