## 1. Go 工作流工程搭建
- [x] 1.1 在项目根目录下创建独立的 Go module（例如 go-audit-workflow）并初始化 go.mod，声明依赖 github.com/cloudwego/eino。
- [x] 1.2 设计目录结构（cmd/workflow、internal/fetch、internal/ai、internal/submit、internal/eino、internal/config），对齐 openspec/project.md 中的约定。
- [x] 1.3 在 internal/config 中实现配置加载逻辑，复用现有 config/app.json 与 config/secrets.local.json 的字段定义与层级结构，在 Go 侧建立与 Python 一致的配置映射。
- [x] 1.4 在 Go 侧支持与 Python 一致的环境变量覆盖策略（如 FETCH_CONFIG、FETCH_SECRETS、YH_CONFIG、YH_SECRETS、AI_PROVIDER、AI_PROMPT_PATH、AI_DEBUG 等），确保配置来源与优先级保持对齐。

## 2. Fetch 节点实现（Go）
- [x] 2.1 在 internal/httpclient 中封装基础 HTTP 客户端，支持 verify_ssl 与超时配置。
- [x] 2.2 在 internal/fetch 中实现登录与列表/详情拉取逻辑，对齐 1-fetch/main.py 的语义与查询参数。
- [x] 2.3 在 Fetch 节点中将结果写入 data/pending_audits.jsonl，保证与 Python 输出格式兼容。

## 3. AI 节点实现（Go + Eino 模型组件）
- [ ] 3.1 基于 Eino 的模型组件在 internal/ai 中封装对 LLM 的调用接口，支持从配置读取 Provider、模型名称与超时等参数。
- [ ] 3.2 在 AI 节点中读取 data/pending_audits.jsonl，构造与 ai/main.py 一致的 prompt 字段，并写入 data/pending_audits_risk.jsonl。
- [ ] 3.3 在 AI 节点中实现风险评分解析与边界检查（1~10），对齐 _parse_score 的行为。
- [ ] 3.4 在 Go 侧复用现有大模型提示词模板，支持 AI_PROMPT_PATH、ai.prompt_path 与内置 ai/prompts/risk.json 的加载优先级，与 Python 的 _load_prompt_template 与 build_risk_prompt 行为保持一致。

## 4. Submit 节点实现（Go）
- [x] 4.1 在 internal/submit 中实现登录与审核结果提交逻辑，对齐 2-submit/main.py 的路径与字段映射。
- [x] 4.2 在 Submit 节点中读取 data/pending_audits_risk.jsonl，构造提交 payload 并处理成功/失败统计。
- [ ] 4.3 为 Submit 节点增加基础重试与错误日志，便于在 Graph 中观测失败原因。

## 5. Eino Graph 编排与运行入口
- [x] 5.1 在 internal/eino 中定义基于 compose.Graph 或 Workflow 的线性工作流，将 Fetch → AI → Submit 作为三个顺序节点。
- [x] 5.2 在 cmd/workflow/main.go 中初始化配置与依赖，构建并编译 Graph，提供单次运行入口。
- [ ] 5.3 确保工作流中每个节点完成后均可以通过 data/ 中间文件进行断点续跑或单独调试。

## 6. 验证与对齐
- [ ] 6.1 使用本地 config/app.json 与 config/secrets.local.json 运行 Go 工作流，与 Python 三段式脚本输出进行对比（记录数量与关键字段）。
- [ ] 6.2 通过实际调用御衡平台验证 Fetch 与 Submit 行为与现有脚本一致。
- [ ] 6.3 更新相关文档或开发说明，记录 Python 与 Go 两套实现的使用方式与迁移建议。
