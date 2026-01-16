# 提案：精简 LLM 入参并补齐 ATT&CK 选择能力

## 摘要
本提案解决两个问题：
- 传给 LLM 的上下文可能过大：引入“可控裁剪”的输入构造策略，只传关键字段（如 PoC/描述/关键请求响应片段），避免把整条 JSON 原样塞进提示词。
- LLM 无法基于完整 ATT&CK 做判断：当前提示词仅内置少量 ATT&CK 片段，且不会使用仓库中的 [ATT&CK.csv](file:///Users/chenshurong/Desktop/Ai-work/sss/ATT%26CK.csv)。本提案引入“基于 ATT&CK.csv 的候选集注入”能力，让 LLM 在每条记录的提示词中从候选列表选择战术/技术/子技术。

## 现状
- LLM 提示词模板位于 [risk.json](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/components/prompts/risk.json)，其中 ATT&CK 约束仅为手写的“部分常用”列表。
- ATT&CK.csv 当前只在提交阶段通过 taxonomy 映射使用（名称→ID），LLM 阶段并不知道完整枚举，因此无法可靠输出完整覆盖的 tactic/technique/sub。

## 目标
1. LLM 入参“可控且可解释”：仅传关键字段，并对单字段长度与总长度做限制。
2. LLM 的 tactic/technique/sub 选择“有依据”：提示词包含由 ATT&CK.csv 生成的候选列表，要求 LLM 只能从候选中选择中文规范名（禁止数值 ID）。
3. 在 token 可控前提下提高准确性：采用“两阶段选择”以降低一次性注入候选集的长度压力，并提升 technique/sub 的命中率。

## 非目标
- 不在本提案阶段改动抓取/提交 API 形态。
- 不强制把完整 ATT&CK.csv 全量塞进提示词（token 成本过高且不可控）。

## 方案概述
### LLM 入参裁剪
为每条漏洞记录构造“精简上下文”，按优先级拼装并截断：
- 漏洞名称、漏洞描述
- PoC/验证证据（如存在）
- 关键请求/响应片段（仅首/尾若干行或固定字节数）
- 其他必要字段（如影响面、攻击类型提示）

### ATT&CK 两阶段选择（推荐）
从 ATT&CK.csv 加载完整枚举，但不全量注入提示词；改为两阶段交互：
1) 第一阶段：仅注入“战术候选”（例如：侦察/资源开发/初始访问…），要求 LLM 只选择 tactic_name（中文规范名，禁止数值 ID）。
2) 第二阶段：根据第一阶段选出的 tactic_name，仅注入该战术下的 technique/sub_technique 候选（并受 Top-K 与长度预算限制），要求 LLM 输出 technique_name 与 sub_technique_name（中文规范名，禁止数值 ID）。

两阶段交互的输出合并后，形成最终的 tactic/technique/sub 结构化字段。

候选集生成策略（实现阶段落地）：
- 第一阶段候选：按配置列出若干战术（可按业务常见范围缩减，比如优先“侦察/资源开发/初始访问”等）。
- 第二阶段候选：在所选战术范围内，基于名称/描述/PoC 关键词做轻量匹配或模糊匹配，产出 Top-K technique/sub 候选。

### 备选方案：单阶段候选集注入（未选）
一次性注入“tactic + technique + sub”候选子集并要求 LLM 同时选择三者。该方案实现简单但更容易触碰长度预算，且 technique/sub 的候选约束更难做到既全面又可控。
