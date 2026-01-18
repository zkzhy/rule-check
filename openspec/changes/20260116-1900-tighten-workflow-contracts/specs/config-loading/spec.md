# Spec: Config Loading Error Policy

## ADDED Requirements

### Requirement: secrets 文件存在但非法必须失败
当 `YH_SECRETS` 指向的文件存在但不可读或不是合法 JSON 时，配置加载 MUST 失败并返回错误。

#### Scenario: secrets 文件 JSON 非法
- **Given** `YH_SECRETS` 指向一个存在但内容非法的文件
- **When** 调用配置加载
- **Then** 返回错误而不是吞错继续

### Requirement: secrets 文件不存在允许为空
当 `YH_SECRETS` 指向的文件不存在时，配置加载 MUST 继续执行，并将 secrets 视为空对象。

#### Scenario: secrets 文件缺失
- **Given** `YH_SECRETS` 指向一个不存在的文件
- **When** 调用配置加载
- **Then** 配置加载成功

## MODIFIED Requirements

## REMOVED Requirements

