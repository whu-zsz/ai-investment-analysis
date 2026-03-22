# AI投顾助手 · API Interface

本文档面向后端开发与前后端联调，重点约定 `frontend2` 当前需要的接口、请求参数、返回字段和状态码建议。

## 1. 通用约定

### 1.1 Base URL

开发环境建议：

```text
/api/v1
```

示例：

```text
http://localhost:8080/api/v1
```

### 1.2 通用请求头

登录前接口：

```http
Content-Type: application/json
```

登录后接口建议：

```http
Content-Type: application/json
Authorization: Bearer <token>
```

### 1.3 通用响应结构

建议所有接口统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| code | number | 业务状态码，`0` 表示成功 |
| message | string | 响应描述 |
| data | object \| array \| null | 具体业务数据 |

### 1.4 通用分页结构

列表接口建议：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [],
    "page": 1,
    "pageSize": 10,
    "total": 100
  }
}
```

### 1.5 通用状态码建议

| code | 含义 |
|------|------|
| 0 | 成功 |
| 40001 | 请求参数错误 |
| 40002 | 文件格式不支持 |
| 40003 | 字段映射校验失败 |
| 40101 | 未登录或 token 无效 |
| 40301 | 无权限访问 |
| 40401 | 资源不存在 |
| 42901 | AI 接口调用过于频繁 |
| 50001 | 服务内部错误 |
| 50002 | AI 分析失败 |

## 2. 认证模块

## 2.1 登录

### 接口

```http
POST /auth/login
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名或学号 |
| password | string | 是 | 登录密码 |

### 请求示例

```json
{
  "username": "盛哲",
  "password": "123456"
}
```

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.token | string | JWT token |
| data.user.id | number | 用户 ID |
| data.user.username | string | 用户名 |
| data.user.email | string | 邮箱 |
| data.user.avatarUrl | string | 头像地址 |
| data.user.riskTolerance | string | 风险承受能力 |
| data.user.investmentPreference | string | 投资偏好 |

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "jwt-token-demo",
    "user": {
      "id": 1,
      "username": "盛哲",
      "email": "demo@example.com",
      "avatarUrl": "",
      "riskTolerance": "medium",
      "investmentPreference": "balanced"
    }
  }
}
```

## 2.2 获取当前用户信息

### 接口

```http
GET /auth/me
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.id | number | 用户 ID |
| data.username | string | 用户名 |
| data.email | string | 邮箱 |
| data.phone | string | 手机号 |
| data.avatarUrl | string | 头像 |
| data.totalProfit | number | 累计盈亏 |
| data.riskTolerance | string | 风险承受能力 |
| data.investmentPreference | string | 投资偏好 |

## 2.3 退出登录

### 接口

```http
POST /auth/logout
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data | null | 可返回 null |

## 3. Dashboard 模块

## 3.1 首页汇总数据

### 接口

```http
GET /dashboard/summary
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.totalAsset | number | 总资产 |
| data.profit30d | number | 近 30 日盈亏 |
| data.profit30dRate | number | 近 30 日收益率 |
| data.riskLevel | string | 风险等级 |
| data.analysisCompletion | number | AI 分析完成度，0-100 |

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "totalAsset": 428560,
    "profit30d": 16840,
    "profit30dRate": 3.8,
    "riskLevel": "medium",
    "analysisCompletion": 89
  }
}
```

## 3.2 当前持仓分布

### 接口

```http
GET /dashboard/portfolio
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data[].assetCode | string | 资产代码 |
| data[].assetName | string | 资产名称 |
| data[].assetType | string | 资产类型 |
| data[].holdingRatio | number | 持仓占比 |
| data[].profitRate | number | 收益率 |
| data[].marketValue | number | 当前市值 |

## 3.3 收益趋势与预测

### 接口

```http
GET /dashboard/trend
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 时间维度，默认 `6m` |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.period | string | 时间周期 |
| data.points[].label | string | 横轴标签 |
| data.points[].actual | number | 历史值 |
| data.points[].forecast | number | 预测值 |

## 3.4 风险提醒

### 接口

```http
GET /dashboard/alerts
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data[].id | number | 提醒 ID |
| data[].level | string | 提醒级别 |
| data[].content | string | 提醒内容 |
| data[].createdAt | string | 创建时间 |

## 4. 文件上传模块

## 4.1 上传投资记录文件

### 接口

```http
POST /files/upload
```

### Content-Type

```http
multipart/form-data
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 上传文件 |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.fileId | number | 文件 ID |
| data.fileName | string | 文件名 |
| data.fileSize | number | 文件大小，字节 |
| data.fileType | string | 文件类型 |
| data.status | string | 上传状态 |
| data.uploadedAt | string | 上传时间 |

## 4.2 解析上传文件

### 接口

```http
POST /files/{fileId}/parse
```

### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fileId | number | 是 | 文件 ID |

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.fileId | number | 文件 ID |
| data.detectedFields[].fieldName | string | 检测到的字段名 |
| data.detectedFields[].mappedKey | string | 建议映射字段 |
| data.detectedFields[].status | string | `recognized` / `pending` |
| data.detectedFields[].sampleValue | string | 样例值 |
| data.warnings[] | string | 解析警告信息 |

## 4.3 确认字段映射并入库

### 接口

```http
POST /files/{fileId}/confirm
```

### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fileId | number | 是 | 文件 ID |

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| mappings | array | 是 | 字段映射列表 |

### mappings 子项字段

| 字段名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sourceField | string | 是 | 原始字段名 |
| targetField | string | 是 | 目标字段名 |

### 请求示例

```json
{
  "mappings": [
    {
      "sourceField": "成交日期",
      "targetField": "transaction_date"
    },
    {
      "sourceField": "证券代码",
      "targetField": "asset_code"
    }
  ]
}
```

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.fileId | number | 文件 ID |
| data.importedCount | number | 成功导入条数 |
| data.failedCount | number | 导入失败条数 |
| data.status | string | 导入状态 |

## 4.4 获取上传历史

### 接口

```http
GET /files/history
```

### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页数量，默认 10 |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.list[].fileId | number | 文件 ID |
| data.list[].fileName | string | 文件名 |
| data.list[].status | string | 状态 |
| data.list[].uploadedAt | string | 上传时间 |
| data.list[].importedCount | number | 导入成功数 |

## 5. 交易记录模块

## 5.1 获取交易记录列表

### 接口

```http
GET /transactions
```

### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | number | 否 | 页码 |
| pageSize | number | 否 | 每页条数 |
| startDate | string | 否 | 开始日期 |
| endDate | string | 否 | 结束日期 |
| assetCode | string | 否 | 资产代码 |
| transactionType | string | 否 | `buy` / `sell` / `dividend` |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.list[].id | number | 交易 ID |
| data.list[].transactionDate | string | 交易日期 |
| data.list[].transactionType | string | 交易类型 |
| data.list[].assetCode | string | 资产代码 |
| data.list[].assetName | string | 资产名称 |
| data.list[].quantity | number | 数量 |
| data.list[].pricePerUnit | number | 单价 |
| data.list[].totalAmount | number | 总金额 |
| data.list[].status | string | 状态 |

## 5.2 获取交易详情

### 接口

```http
GET /transactions/{id}
```

### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | 是 | 交易 ID |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.id | number | 交易 ID |
| data.transactionDate | string | 交易日期 |
| data.transactionType | string | 交易类型 |
| data.assetType | string | 资产类型 |
| data.assetCode | string | 资产代码 |
| data.assetName | string | 资产名称 |
| data.quantity | number | 交易数量 |
| data.pricePerUnit | number | 单价 |
| data.totalAmount | number | 总金额 |
| data.commission | number | 手续费 |
| data.profit | number | 盈亏 |
| data.notes | string | 备注 |
| data.sourceFile | string | 来源文件 |

## 6. AI 分析模块

## 6.1 触发分析

### 接口

```http
POST /analysis/generate
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fileId | number | 否 | 基于某次上传生成分析 |
| analysisType | string | 否 | 默认 `full` |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.reportId | number | 分析报告 ID |
| data.status | string | `pending` / `running` / `completed` / `failed` |
| data.startedAt | string | 开始时间 |

## 6.2 获取最新分析结果

### 接口

```http
GET /analysis/latest
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.reportId | number | 报告 ID |
| data.summary | string | 投资总结 |
| data.investmentPreference | string | 投资偏好 |
| data.riskLevel | string | 风险等级 |
| data.behaviorTags[] | string | 行为标签 |
| data.patternRecognition | string | 模式识别结果 |
| data.riskWarnings[] | string | 风险提示 |
| data.createdAt | string | 生成时间 |

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "reportId": 101,
    "summary": "近期交易风格由进攻型向平衡型过渡。",
    "investmentPreference": "balanced",
    "riskLevel": "medium",
    "behaviorTags": ["偏成长", "偏爱ETF", "适合月度复盘"],
    "patternRecognition": "短线频繁调仓胜率偏低，中长期持有收益更稳定。",
    "riskWarnings": [
      "新能源仓位偏高",
      "存在情绪化卖出行为"
    ],
    "createdAt": "2026-03-20T16:40:00Z"
  }
}
```

## 6.3 获取分析历史

### 接口

```http
GET /analysis/history
```

### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | number | 否 | 页码 |
| pageSize | number | 否 | 每页数量 |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.list[].reportId | number | 报告 ID |
| data.list[].title | string | 报告标题 |
| data.list[].riskLevel | string | 风险等级 |
| data.list[].createdAt | string | 创建时间 |
| data.list[].status | string | 状态 |

## 6.4 获取指定分析报告详情

### 接口

```http
GET /analysis/{reportId}
```

### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reportId | number | 是 | 报告 ID |

### 返回字段

建议与 `GET /analysis/latest` 基本一致，可增加：

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.charts | object | 图表数据 |
| data.sourceFileId | number | 来源文件 ID |

## 7. 趋势预测模块

## 7.1 获取最新预测结果

### 接口

```http
GET /prediction/latest
```

### 请求参数

无

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.confidence | number | 置信度 |
| data.points[].label | string | 横轴标签 |
| data.points[].actual | number | 历史值 |
| data.points[].forecast | number | 预测值 |
| data.scenarios[].title | string | 情景标题 |
| data.scenarios[].range | string | 收益区间 |
| data.scenarios[].detail | string | 情景说明 |
| data.generatedAt | string | 生成时间 |

## 7.2 重新生成预测

### 接口

```http
POST /prediction/generate
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 预测周期，如 `30d` |
| model | string | 否 | 使用的模型标识 |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.taskId | number | 任务 ID |
| data.status | string | 任务状态 |

## 8. 报告导出模块

## 8.1 导出 PDF 报告

### 接口

```http
GET /reports/{reportId}/export
```

### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reportId | number | 是 | 报告 ID |

### 返回方式建议

方案一：

- 直接返回 PDF 文件流

方案二：

- 返回下载地址

### 如果返回下载地址，建议响应字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.reportId | number | 报告 ID |
| data.downloadUrl | string | 下载地址 |
| data.expireAt | string | 链接过期时间 |

## 9. 用户设置模块

## 9.1 获取用户设置

### 接口

```http
GET /settings
```

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data.riskPreference | string | 风险偏好 |
| data.defaultAnalysisPeriod | string | 默认分析周期 |
| data.reportTemplate | string | 报告模板 |
| data.enableAlerts | boolean | 是否开启提醒 |

## 9.2 更新用户设置

### 接口

```http
PUT /settings
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|------|------|------|------|
| riskPreference | string | 否 | 风险偏好 |
| defaultAnalysisPeriod | string | 否 | 默认分析周期 |
| reportTemplate | string | 否 | 报告模板 |
| enableAlerts | boolean | 否 | 是否开启提醒 |

### 返回字段

| 字段路径 | 类型 | 说明 |
|------|------|------|
| data | object | 更新后的设置 |

## 10. 联调优先级建议

建议后端开发和联调顺序如下：

1. `POST /auth/login`
2. `GET /auth/me`
3. `GET /dashboard/summary`
4. `GET /dashboard/portfolio`
5. `GET /dashboard/alerts`
6. `POST /files/upload`
7. `POST /files/{fileId}/parse`
8. `POST /analysis/generate`
9. `GET /analysis/latest`
10. `GET /prediction/latest`
11. `GET /transactions`
12. `GET /reports/{reportId}/export`

## 11. 推荐补充

为了让前端更顺利接入，建议后端额外统一以下细节：

- 时间字段统一使用 ISO 8601 字符串
- 金额字段统一返回 number，不混用字符串
- 枚举字段尽量固定，如 `buy`、`sell`、`dividend`
- 文件上传和 AI 任务建议返回明确状态字段
- 错误响应也保持统一结构，方便前端统一提示
