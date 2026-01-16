1. 创建 `internal/taxonomy`，加载 `/Users/chenshurong/Desktop/Ai-work/sss/ATT&CK.csv`，提供名称→ID 映射。
2. 更新 `ai/prompts/risk.json`，只要求名称字段（tactic_name/technique_name/sub_technique_name），禁止输出数值 ID。
3. 更新 `internal/llm/llm.go`，解析名称-only 字段并存入 `data`。
4. 更新 `internal/submit/submit.go`，在提交阶段通过 taxonomy 映射构造 `tactics` 对象（含 ID 与中文名称）。
5. 校验 `devices`：优先 `_raw.devices` 的 `{id,name}` 对象数组；形态非法则不提交该字段以避免 400。
6. 保证 `attack_result="成功"`、`score` 规范到 (0,10]、`level_id` 为 1|2|3。
7. 增加单测覆盖 taxonomy 映射与完整载荷构造；DRY_RUN 输出与示例对比。
8. 端到端提交验证返回 200；名称不匹配时降级仅剧情并告警。
