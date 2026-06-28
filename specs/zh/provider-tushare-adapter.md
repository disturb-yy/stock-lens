# Tushare Provider 适配器

本文档冻结第一阶段 Tushare provider 适配规则。

## 边界

原始 Tushare HTTP client 位于 `pkg/tushare`。它负责 Tushare 协议细节，例如 token 处理、`api_name`、显式 `fields`、原始 request/response 结构、provider 错误码、限流和可选的底层重试。

Market 领域 Tushare 适配器位于 `domain/market/provider_tushare.go`。它将原始 Tushare 数据转换为 market 领域模型，并通过 provider 接口返回这些模型。

Service 不处理原始 Tushare response 结构。

## 身份映射

`stocks.ts_code` 存储 Tushare 标识，例如 `600519.SH`。

`daily_k_lines` 不存储 `ts_code`。

Provider 从 `ts_code` 提取系统 `symbol`，例如 `600519.SH` -> `600519`。无效 `ts_code` 是 provider mapping error。

Exchange 映射使用 `ts_code` 后缀或 Tushare exchange 字段：

```text
.SH -> SSE
.SZ -> SZSE
.BJ -> BSE
```

无法识别的 exchange 对股票主数据而言是 provider mapping error。

## 股票映射

Board 推导遵循第一阶段规则：

```text
688xxx          -> STAR
300xxx          -> GEM
8xxxxx / 4xxxxx -> BSE
other           -> MAIN
unknown         -> UNKNOWN
```

Stock status 映射：

```text
L -> LISTED
D -> DELISTED
P -> PAUSED
```

未知 stock status 是 provider mapping error。

Tushare provider 输出使用 `DataSource = TUSHARE`。

## 日期和 Decimal 映射

在 Tushare 边界，日期使用 `YYYYMMDD`。

领域模型使用 date/time 类型。HTTP API 使用 `YYYY-MM-DD`。MySQL 将交易日期存储为 `DATE`。

Tushare 数值转换为 `decimal.Decimal`。空字符串、非数字值和精度解析失败都是 provider mapping error。

存储的市场数据单位保持 Tushare 原始单位：

```text
volume     = lot
amount     = thousand_cny
pct_change = percent
```

## 日 K

第一阶段只获取并存储原始、不复权日 K。

Provider 不得请求前复权或后复权数据。如果 Tushare API 要求复权参数，provider 必须显式请求原始不复权数据。

## 空结果

Provider 空结果处理遵循同步任务行为：

- empty stock list：`MARKET_PROVIDER_ERROR`
- empty trade calendar：`MARKET_PROVIDER_ERROR`
- empty single-stock daily K line result：成功，零行
- full-market daily K sync 中某只股票空结果：股票级成功，零行

## 错误处理

Provider 错误包装为 market 领域错误，通常是 `MARKET_PROVIDER_ERROR`。

内部错误文本可以写入同步日志和应用日志，但不得记录 Tushare token 等敏感值。

Tushare token 值、包含 token 的 request header，以及包含 token 的完整 request body 都不得出现在日志中。

## 测试和重试

第一阶段不运行真实 Tushare 网络测试。

Provider 测试使用构造的原始响应，而不是真实 Tushare 调用。

第一阶段任务语义不依赖自动重试。如果实现轻量重试，它应位于 `pkg/tushare` 内部，并且不得改变同步任务状态语义。

Raw client 请求应显式列出 Tushare `fields`，避免 Tushare 默认字段变化破坏解析。

## Mock Provider

Mock provider 返回 market 领域模型，并使用 `DataSource = MOCK`。

Mock provider 应支持可控失败场景，以便测试覆盖同步任务失败和部分成功行为。
