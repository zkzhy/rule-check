---
change-id: 20260115-1105-llm-blueview-tactics-devices-ids
summary: 将 LLM 风险输出与蓝队视角提交流程对齐，使用 ATT&CK.csv 的中文规范名称并在提交流构造对应 ID，设备字段为对象数组。
status: proposal
---

问题
- 设备（devices）字段类型不匹配可能导致 400（字符串 vs 对象数组 {id,name}）。
- LLM 输出的“剧情/技术/子技术”未被强约束，后端期望结构化对象并且名称需可映射到 ID。
- 提交载荷需与您提供的 UI/网络请求示例完全一致：devices 为对象数组；tactics 为含 ID 与中文名称的对象；风险字段齐备。

目标
- 约束 LLM 只输出中文规范名称（与 ATT&CK.csv 对应），不输出数值 ID。
- 运行时加载 ATT&CK.csv，将中文名称映射为 ID，在提交阶段构造 `tactics` 对象。
- `devices` 必须是对象数组 `{id, name}`；优先使用 `_raw.devices` 并校验形态。
- 风险必填字段保证完整：`eval_description`、`suggestion`、`score(1..10)`、`level_id(1/2/3)`、`attack_result="成功"`。

非目标
- 不要求模型记忆或输出数值 ID。
- 不改动服务端 API。

范围
- 涉及文件：`ai/prompts/risk.json`、`internal/llm/llm.go`、`internal/submit/submit.go`、新增 `internal/taxonomy` 模块加载/校验 ATT&CK.csv。
- 数据来源：`/Users/chenshurong/Desktop/Ai-work/sss/ATT&CK.csv`。

方案
1) 提示词调整
   - 明确 LLM 仅输出中文规范名称（与 ATT&CK.csv 对齐），JSON 键使用以下命名：
     - `tactic_name`（所属剧情）、`technique_name`（对应技术）、`sub_technique_name`（对应子技术，允许为空）
     - LLM 不输出 ID；ID 在提交阶段由程序依据 ATT&CK.csv 映射得到。
   - 保留风险字段：`eval_description`、`suggestion`、`risk_score(1..10)`、`level_id(1/2/3)`。
   - 不要求 LLM 输出 `devices`；管道会复用 `_raw.devices`（若存在）并进行形态校验。

2) 分类表加载器
   - 新增 `internal/taxonomy`：
     - 加载 ATT&CK.csv，按中文名称建立内存索引。
     - 提供 `LookupIDs(tacticName, techniqueName, subName) -> (tacticID, techniqueID, subID)`。
     - 校验名称、规范化空白；子技术为空时返回 0。

3) 管道映射
   - llm.go：解析名称-only 字段到 `data["tactic_name"]`、`data["technique_name"]`、`data["sub_technique_name"]`。
   - submit.go：
     - 通过 taxonomy 映射构造 `tactics: [{tactic_id, tactic_name, technique_id, technique_name, sub_technique_id, sub_technique_name}]`。
     - `attack_result` 无条件设为 `"成功"`。
     - `devices`：优先使用 `_raw.devices`（形态为 `[{id,name}]`）；若不存在且 LLM 提供对象数组则校验通过后使用；否则为避免 400 不提交该字段。
     - `score` 规范化到 (0,10]；`level_id` 直接使用 LLM 给出值（1|2|3）。

验证
- 单元测试：taxonomy 映射覆盖与构造校验。
- 端到端干跑（`DRY_RUN=1`）：读取示例记录，输出与您提供的载荷（ID 14748）完全一致的 JSON。
- 实际提交：期望 200；若返回 400，则明确日志指出失败字段。

风险与缓解
- 名称不匹配：进行规范化与告警；若技术/子技术查找失败，退化为仅提交剧情（含 ID 与名称）。
- 设备缺失或类型不对：不提交 devices 以避免类型错误，同时记录警告。

结果
- 提交载荷与 UI/示例一致，消除类型错误，并使用 ATT&CK.csv 映射保证剧情/技术/子技术 ID 与中文名称正确。
