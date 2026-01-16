## 修改的需求

### 需求：LLM 风险输出必须使用 ATT&CK.csv 的中文规范名称
- 输出包含：
  - eval_description: 字符串
  - suggestion: 字符串
  - risk_score: 整数（1..10）
  - level_id: 整数（1|2|3）
  - tactic_name: 字符串（来自 ATT&CK.csv）
  - technique_name: 字符串（来自 ATT&CK.csv）
  - sub_technique_name: 字符串或空（来自 ATT&CK.csv）
- 不得输出数值 ID；ID 由管道在提交阶段使用 ATT&CK.csv 映射。

#### 场景：仅名称输出被解析并接受
- 给定待评估记录（含名称与描述）
- 当模型返回仅包含名称的字段
- 则管道将其存入 `data` 并在提交阶段进行映射

## 新增的需求

### 需求：名称需与 ATT&CK.csv 校验一致
- 管道必须校验返回名称在 ATT&CK.csv 中存在。
- 若技术/子技术名称不匹配，则仅提交剧情（含剧情 ID 与名称，技术/子技术 ID 省略或为 0），并记录警告。

#### 场景：技术名称不匹配
- 给定 `tactic_name=侦察`、`technique_name=主动扫描X`
- 当映射查找失败
- 则仅提交剧情信息并告警
