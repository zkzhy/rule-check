# 任务：新增 Chaitin OpenAI-Compatible ChatModel 并设为默认

1. 增加 `ai.provider=chaitin` 的 provider 别名
   - 在模型工厂中将 `chaitin` 映射到现有 OpenAI-Compatible 实现
   - 保持现有 provider 行为不变
2. 调整默认配置以指向 Chaitin
   - 更新 `config/app.json` 的 `ai.provider/ai.base_url/ai.model`
   - 确保敏感字段仍只从 `config/secrets.local.json` 读取
3. 增加验证用单测
   - 覆盖 base_url 拼接规则（带/不带 `/v1`）
   - 覆盖 provider 选择逻辑（`chaitin` 分支命中）
4. 验证与工具链
   - 运行 `go test ./...`、`go build ./...`、`go vet ./...`

