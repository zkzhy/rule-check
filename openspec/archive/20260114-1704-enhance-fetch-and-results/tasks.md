# 任务清单：配置化 Fetch 与结构化结果拼接

1. 为 fetch 增加配置字段（endpoint、method、send_style、page_size、filters、time_fields）。
2. 实现请求构造：支持 GET query 或 POST JSON body。
3. 完成分页终止条件：返回数量 < page_size 或空列表即停止。
4. 增加调试日志（脱敏）：方法、URL、payload 概要。
5. 修改不同 filters 验证返回结果会变化（不再固定）。
6. 为 LLM 引入结构化输出提示词与 JSON 解析。
7. 生成 `pending_audits_results.jsonl`，按“缺失补全”方式拼接字段。
8. 对 risk_score 做范围裁剪（0..10），并复用上游已有 risk_score。
9. 从可配置文件读取枚举映射（路径 + 格式），强制所属剧情/对应技术/标签一一对应。
10. 对已存在的 suggestion 字段：不推送给 LLM，输出直接复用原值。
11. 在 project.md 中补充配置示例（仅说明，不写实现代码）。
12. 端到端验证：go run cmd/workflow/main.go 并检查输出文件。
