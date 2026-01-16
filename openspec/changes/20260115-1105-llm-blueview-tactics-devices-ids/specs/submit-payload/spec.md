## 修改的需求

### 需求：提交载荷必须包含带 ID 的结构化剧情对象
- 构造 `tactics` 为数组（至少一个对象）：
  - tactic_id, tactic_name
  - technique_id, technique_name
  - sub_technique_id, sub_technique_name（无子技术时 sub_technique_id=0，sub_technique_name 为空）
- ID 由名称通过 ATT&CK.csv 映射获得。

#### 场景：完整映射成功
- 给定名称在 ATT&CK.csv 中有效
- 则提交载荷包含所有 ID 与中文名称，与提供示例一致

### 需求：设备（devices）必须为对象数组
- 当存在 `_raw.devices` 且形态为 `[{id,name},...]` 时使用之。
- 若 LLM 提供设备数组，只有在形态合法时才采用；否则不提交该字段。

#### 场景：设备类型不匹配
- 给定 devices 为字符串
- 则不在载荷中包含 devices，并记录警告

## 新增的需求

### 需求：攻击结果与风险字段
- 始终设置 `attack_result="成功"`。
- 将 `score` 规范化到 (0,10]；`level_id` 取值 1|2|3。
- 保留 `eval_description`、`suggestion`。

#### 场景：示例 ID 14748 的提交
- 给定示例记录
- 则最终提交载荷与提供的 JSON 相等，并成功提交（200）
