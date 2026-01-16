# go-audit-workflow 16.18.32（归档说明）

## 这是什么
一个纯 Go 的“漏洞自动审核”工作流：从御衡平台抓取待审核漏洞记录，调用大模型生成风险评分与 ATT&CK 归属（两阶段候选集），再把审核结果回写到平台。

核心目标：
- 可批量处理“待审核”的 HTTP 漏洞记录
- LLM 入参尽量小（只传关键字段），减少模型拒答/超长失败
- ATT&CK 不把整张 CSV 塞进模型（分两次候选注入）
- 全流程可断点续跑（基于 JSONL 中间文件）

## 目录结构
- cmd/workflow/main.go：命令入口（支持按阶段运行）
- internal/fetch：登录、列表分页、详情抓取，输出 JSONL
- internal/orchestrator：AI 风险分析与工作流编排
- internal/components/model：模型 Provider 适配（OpenAI-Compatible 等）
- internal/components/tools/taxonomy：ATT&CK.csv 加载与候选生成/映射
- internal/components/tools/submit：回写审核结果
- internal/httpclient：HTTP JSON 解码与错误增强
- internal/config：配置加载（app + secrets）与默认值
- data/*.jsonl：运行中间文件

## 数据文件（JSONL）
- data/pending_audits.jsonl：Fetch 输出（原始详情 + 精简字段）
- data/pending_audits_results.jsonl：AI 输出（风险分 + tactic/technique/sub + 其他结构化字段），Submit 读取它回写平台

## 快速开始
前置：
- Go 1.21+
- 可访问御衡平台 API
- 准备好 config/app.json 与 config/secrets.local.json（敏感信息仅放 secrets）

### 1) 配置
config/app.json（非敏感）最小示例：
```json
{
  "paths": {
    "output_file": "data/pending_audits.jsonl"
  },
  "yuheng": {
    "base_url": "https://yhope.pl.in.chaitin.net",
    "verify_ssl": false,
    "timeout_s": 20
  },
  "ai": {
    "provider": "chaitin",
    "model": "gpt-5.1",
    "timeout_s": 120,
    "base_url": "https://aiapi.chaitin.net/v1"
  }
}
```

config/secrets.local.json（敏感，本地文件）示例：
```json
{
  "yuheng": {
    "username": "YOUR_USER",
    "password": "YOUR_PASSWORD"
  },
  "ai": {
    "api_keys": {
      "chaitin": "sk-..."
    }
  }
}
```

配置查找顺序（默认）：
- YH_CONFIG（默认 config/app.json）
- YH_SECRETS（默认 config/secrets.local.json）

### 2) 运行方式（按阶段）
全流程（会重新抓取）：
```bash
go run cmd/workflow/main.go -mode=full
```

仅抓取：
```bash
go run cmd/workflow/main.go -mode=fetch
```

按阶段跑（推荐大批量 7000+ 场景）：
- 只跑 AI（基于已存在的 data/pending_audits.jsonl，不重新抓取）：
```bash
go run cmd/workflow/main.go -mode=ai
```

- 只跑 Submit（基于已存在的 data/pending_audits_results.jsonl）：
```bash
go run cmd/workflow/main.go -mode=submit
```

- 跳过 Fetch，直接 AI+Submit（基于已抓取数据，不重新抓取）：
```bash
go run cmd/workflow/main.go -mode=ai-submit
```

断点续跑（AI 阶段）：
```bash
go run cmd/workflow/main.go -mode=ai-submit -resume-ai=true
```
行为：
- 读取 data/pending_audits_results.jsonl 中已存在的 id
- AI 阶段跳过这些 id
- 结果文件以追加方式写入

并发与限速（AI 阶段）：
- 目前命令行参数里没有 `-concurrency` 之类的 flag
- 你可以通过「配置文件」或「环境变量」设置并发与请求速率

方式 A：环境变量（临时生效，推荐快速切换）
```bash
AI_CONCURRENCY=8 AI_RATE_LIMIT_QPS=4 go run cmd/workflow/main.go -mode=ai-submit -resume-ai=true
```

方式 B：config/app.json（持久生效）
在 `ai` 下增加：
```json
{
  "ai": {
    "concurrency": 8,
    "rate_limit_qps": 4
  }
}
```

## Fetch：查询过滤与调试
Fetch 的列表查询支持配置化：
- yuheng.list_endpoint（默认 /api/lines/operation）
- yuheng.list_method（GET/POST，默认 GET）
- yuheng.list_send_style（query/json，默认 query）
- yuheng.list_page_size（默认 1000）
- yuheng.list_filters（自定义过滤条件）
- yuheng.list_time_fields（时间字段键名映射）

默认行为：
- 未配置 review_status 时默认为 “待审核”
- 未配置 type 时默认为 “HTTP”

调试：
```bash
FETCH_DEBUG=1 go run cmd/workflow/main.go -mode=full
```

## AI：精简入参 + ATT&CK 两阶段选择
### 入参裁剪（Context Trimming）
AI 只会把这些字段拼成 context（并按预算截断）：
- name（漏洞名称）
- description（漏洞描述）
- xray_poc_content（PoC/证据）
- req_pkg（请求包）
- resp_pkg（响应包）

预算配置在 ai.context.*（有默认值）。

### 并发与限速（Concurrency / Rate Limit）
AI 支持并发处理与调用限速，配置项在 `ai` 下：
- ai.concurrency：并发 worker 数（默认 1）
- ai.rate_limit_qps：每秒最多请求数（默认 0，表示不额外限速；仍保留轻微调用间隔）

### ATT&CK 两阶段候选注入
目标：不把整个 ATT&CK.csv 传给模型。

流程：
1. 第一阶段：只给 tactic 候选列表，让模型选出 tactic_name
2. 第二阶段：只给所选 tactic 下的 technique/sub 候选（Top-K + 长度预算），让模型选 technique_name/sub_technique_name
3. 输出校验：technique/sub 不命中候选则清空，tactic 保底回退到候选第一项

ATT&CK.csv 路径：
- ai.attck.csv_path（推荐显式配置）
- 或者放在 ./ATT&CK.csv / ../ATT&CK.csv

## Submit：回写规则
Submit 读取 data/pending_audits_results.jsonl：
- 取 risk_score（1..10）
- 取 tactic/technique/sub 的中文名称并映射成平台所需的 ID
- 组装 payload 调用御衡审核接口写回

## 常见问题（排障）
### 1) decode json failed / invalid character '<'
一般是 base_url 配成了网页登录地址或命中了重定向，返回 HTML 不是 JSON。
建议：
- yuheng.base_url 使用 API 根，例如 https://yhope.pl.in.chaitin.net

### 2) 只处理了 213 条（或某个固定数）就停了
典型原因是 JSONL 单行太长，bufio.Scanner 默认 64KB 上限导致后续行读不到。
本项目已把读取 JSONL 的上限提高到 16MB/行（AI 与 Submit 都已处理）。

## 代码入口索引
- 入口与参数解析：[main.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/cmd/workflow/main.go)
- Fetch：[fetch.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/fetch/fetch.go)
- AI RiskAnalysis：[risk_analysis.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/orchestrator/risk_analysis.go)
- Submit：[submit.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/components/tools/submit/submit.go)
- Taxonomy：[taxonomy.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/components/tools/taxonomy/taxonomy.go)
- HTTP Client：[httpclient.go](file:///Users/chenshurong/Desktop/Ai-work/sss/go-audit-workflow%2016.18.32/internal/httpclient/httpclient.go)
