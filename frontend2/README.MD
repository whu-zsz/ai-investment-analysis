# AI投顾助手 · frontend2 使用说明

`frontend2` 是根据项目需求文档重新设计的一版前端 UI demo，主要用于课程展示、页面原型确认，以及后续前后端联调的基础界面。

当前版本包含：

- 登录页
- 首页总览 Dashboard
- 投资记录上传页
- AI 分析报告页
- 趋势预测页
- 历史记录页
- 基础路由保护

## 1. 技术栈

- React 19
- TypeScript
- Vite
- react-router-dom
- react-hook-form
- 自定义 CSS

说明：

- 当前 `frontend2` 虽然保留了 `axios`、`echarts-for-react`、`react-dropzone`、`jspdf` 等依赖，但这一版 demo 主要还是静态 UI 原型，尚未真正接入后端接口和真实图表。

## 2. 目录结构

```bash
frontend2/
├── public/
├── src/
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Layout.tsx
│   │   │   ├── Navbar.tsx
│   │   │   └── Sidebar.tsx
│   │   ├── ui/
│   │   │   └── UploadZone.tsx
│   │   └── ProtectedRoute.tsx
│   ├── data/
│   │   └── mock.ts
│   ├── pages/
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Upload.tsx
│   │   ├── Analysis.tsx
│   │   ├── Prediction.tsx
│   │   └── History.tsx
│   ├── router/
│   │   └── index.tsx
│   ├── types/
│   │   └── index.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
└── README.MD
```

## 3. 如何启动

在项目根目录下进入 `frontend2`：

```bash
cd /Users/lnm/Downloads/stock_whu/ai-investment-analysis/frontend2
```

安装依赖：

```bash
npm install
```

启动开发环境：

```bash
npm run dev
```

默认访问地址一般为：

```bash
http://localhost:5173
```

生产构建：

```bash
npm run build
```

本地预览构建结果：

```bash
npm run preview
```

## 4. 当前页面说明

### 4.1 登录页

- 路径：`/login`
- 当前为前端模拟登录
- 登录成功后会在 `localStorage` 中写入演示 token
- 默认演示账号可直接输入任意有效格式账号，界面文案推荐使用：
  - 用户名：`盛哲`
  - 密码：`123456`

### 4.2 首页 Dashboard

展示内容：

- 总资产、近 30 日盈亏、风险等级、分析完成度
- 收益趋势与预测对比的可视化占位
- 当前持仓分布
- 风险提醒

适合作为后续接入以下内容的首页容器：

- 用户资产统计
- 风险评分
- 组合收益曲线
- AI 每日提醒

### 4.3 上传页

展示内容：

- 上传区域
- 文件名展示
- 字段识别结果
- 上传后处理流程

后续可接入：

- 真实文件上传
- 上传进度条
- 字段映射确认
- 数据清洗结果预览

### 4.4 AI 分析页

展示内容：

- 投资偏好画像
- 行为模式识别
- 风险预警
- 行为标签

后续可接入：

- 大模型分析结果
- 文本总结
- 风险评分卡片
- 图表模块

### 4.5 趋势预测页

展示内容：

- 历史与预测趋势对比占位图
- 三种收益情景卡片

后续可接入：

- ECharts 折线图
- 区间预测图
- 不同模型的预测结果对比

### 4.6 历史记录页

展示内容：

- 历史交易表格
- 状态筛选标签

后续可接入：

- 上传批次记录
- 报告生成历史
- 导出记录

## 5. 当前前端实现逻辑

### 5.1 登录态

当前使用前端本地模拟登录：

- `localStorage.setItem('token', 'frontend2-demo-token')`
- `localStorage.setItem('username', values.username.trim())`

路由守卫文件：

- `src/components/ProtectedRoute.tsx`

未登录访问业务页时会自动跳转到 `/login`。

### 5.2 页面数据

当前页面中展示的数据主要来自：

- `src/data/mock.ts`

这一文件里存放的是 demo 用的静态 mock 数据。  
后续接入后端时，可以逐步将这些数据改成接口返回值。

### 5.3 样式

当前页面样式主要集中在：

- `src/index.css`

这版没有依赖现成组件库，样式是手写的，优点是结构清晰、便于自由调整答辩展示效果。

## 6. 建议后端接入的 API 列表

下面是根据当前页面结构整理的一版建议接口清单，便于你们后端按模块对接。

## 6.1 认证模块

### `POST /api/v1/auth/login`

用途：

- 用户登录

前端请求示例：

```json
{
  "username": "盛哲",
  "password": "123456"
}
```

建议响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": 1,
      "username": "盛哲",
      "email": "demo@example.com",
      "riskTolerance": "medium"
    }
  }
}
```

### `GET /api/v1/auth/me`

用途：

- 获取当前登录用户信息

### `POST /api/v1/auth/logout`

用途：

- 退出登录

## 6.2 首页 Dashboard 模块

### `GET /api/v1/dashboard/summary`

用途：

- 获取首页顶部统计数据

建议返回字段：

- 总资产
- 近 30 日盈亏
- 风险等级
- AI 分析完成度

### `GET /api/v1/dashboard/portfolio`

用途：

- 获取当前持仓分布

建议返回字段：

- 资产名称
- 资产代码
- 占比
- 收益率

### `GET /api/v1/dashboard/trend`

用途：

- 获取收益趋势和预测趋势数据

建议返回字段：

- 时间维度
- 历史收益值
- 预测收益值

### `GET /api/v1/dashboard/alerts`

用途：

- 获取 AI 风险提醒

## 6.3 文件上传模块

### `POST /api/v1/files/upload`

用途：

- 上传 CSV / Excel 文件

请求方式：

- `multipart/form-data`

建议返回字段：

- 文件 ID
- 文件名
- 文件大小
- 上传状态

### `POST /api/v1/files/{fileId}/parse`

用途：

- 解析上传文件

建议返回字段：

- 识别出的字段列表
- 字段映射建议
- 异常数据提示

### `POST /api/v1/files/{fileId}/confirm`

用途：

- 用户确认字段映射后入库

### `GET /api/v1/files/history`

用途：

- 获取上传历史

## 6.4 交易记录模块

### `GET /api/v1/transactions`

用途：

- 获取交易记录列表

建议支持查询参数：

- `page`
- `pageSize`
- `startDate`
- `endDate`
- `assetCode`
- `transactionType`

### `GET /api/v1/transactions/{id}`

用途：

- 获取单条交易详情

### `POST /api/v1/transactions`

用途：

- 新增交易记录

### `PUT /api/v1/transactions/{id}`

用途：

- 修改交易记录

### `DELETE /api/v1/transactions/{id}`

用途：

- 删除交易记录

## 6.5 AI 分析模块

### `POST /api/v1/analysis/generate`

用途：

- 触发一次新的 AI 分析

建议请求体：

```json
{
  "fileId": 12,
  "analysisType": "full"
}
```

### `GET /api/v1/analysis/latest`

用途：

- 获取最近一次 AI 分析结果

建议返回字段：

- 投资总结
- 投资偏好
- 风险等级
- 风险提示
- 行为标签
- 模式识别结果

### `GET /api/v1/analysis/history`

用途：

- 获取历史分析报告列表

### `GET /api/v1/analysis/{reportId}`

用途：

- 获取指定分析报告详情

## 6.6 趋势预测模块

### `GET /api/v1/prediction/latest`

用途：

- 获取最新趋势预测结果

建议返回字段：

- 时间序列
- 历史值
- 预测值
- 置信度
- 情景分析列表

### `POST /api/v1/prediction/generate`

用途：

- 重新生成一次趋势预测

## 6.7 PDF 导出模块

### `GET /api/v1/reports/{reportId}/export`

用途：

- 导出 PDF 报告

建议返回：

- 直接返回 PDF 文件流
- 或返回可下载链接

## 6.8 用户设置模块

### `GET /api/v1/settings`

用途：

- 获取用户设置

### `PUT /api/v1/settings`

用途：

- 更新用户设置

建议字段：

- 风险偏好
- 默认分析周期
- 报告导出模板
- 是否开启提醒

## 7. 前后端联调建议

建议联调顺序：

1. 先接登录接口和用户信息接口
2. 再接首页 summary、portfolio、alerts
3. 然后接文件上传和交易记录列表
4. 最后接 AI 分析、趋势预测和 PDF 导出

这样做的原因是：

- 可以先跑通登录态和路由保护
- 首页最快能看到动态数据
- 上传和交易记录是后续 AI 分析的基础
- 分析和预测模块通常依赖更多后端逻辑

## 8. 后续前端建议改造点

- 新增 `src/services/api.ts` 统一封装 axios 实例
- 新增 `src/services/modules/` 存放各模块接口
- 新增 `src/types/api.ts` 管理接口返回类型
- 将 `src/data/mock.ts` 分步替换成真实接口数据
- 接入 `echarts-for-react` 替换当前图表占位
- 接入 `react-dropzone` 实现拖拽上传
- 接入真实 JWT 鉴权和 token 过期处理

## 9. 补充说明

当前 `frontend2` 的定位是：

- 优先保证页面完整
- 优先保证答辩展示效果
- 优先保证后续可扩展性

它还不是最终生产版前端，但已经适合作为你们项目后续联调和继续开发的基础版本。
