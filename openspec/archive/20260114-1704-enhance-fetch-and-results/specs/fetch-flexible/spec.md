## ADDED Requirements

### Requirement: Config-driven fetch filters
系统 MUST 支持通过配置文件和环境变量定义查询过滤项、请求方法和发送样式，以便未来扩展不需要修改代码。

#### Scenario: GET with query parameters
- 给定 list_method=GET 且 list_send_style=query
- 且 list_filters 包含 { name:"foo", review_status:"待审核", type:"HTTP" }
- 当调用 fetch 列表查询
- 则请求 URL 必须包含对应的 query 参数
- 且返回结果会随过滤条件变化（不再固定为同一批数据）

#### Scenario: POST with JSON body
- 给定 list_method=POST 且 list_send_style=json
- 且 list_filters 包含多个字段（包含时间范围）
- 当调用 fetch 列表查询
- 则 HTTP body 必须是 JSON 对象，包含 filters、page_no、page_size
- 且返回结果会随过滤条件变化

#### Scenario: Pagination until exhaustion
- 给定 page_size=N
- 当某一页返回数量 < N
- 则分页必须停止
- 且汇总日志包含总条数与输出文件名

#### Scenario: Debug logging for troubleshooting
- 给定 FETCH_DEBUG=true
- 当执行 fetch
- 则日志输出方法、endpoint、脱敏后的 payload 概要
- 且不会打印任何敏感信息

### Requirement: Time range field mapping
系统 MUST 允许配置时间范围字段的键名映射，以适配后端要求。

#### Scenario: Custom time keys
- 给定 list_time_fields 提供 start_time/end_time/update_start/update_end 的键名映射
- 且 list_filters 提供对应时间范围值
- 当构建请求
- 则最终 query/body 使用映射后的键名
