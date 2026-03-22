```markdown
# AI 投顾助手 · 前端部分 (Ant Design 版)

基于 **React + TypeScript + Vite + Ant Design** 构建的智能投资记录分析与预测 Web 应用。

## 🚀 项目概述

- **目标**：为个人投资者提供“数据导入 → AI 智能分析 → 风险评估 → 趋势预测”的一站式 Web 体验。
- **核心定位**：将复杂的投资交易流水，通过 DeepSeek / GPT 等大模型转化为易读的投资报告与可视化图表。

## 🛠 技术栈更新

为了提升企业级 UI 的开发效率与交互体验，本项目已全面转向 **Ant Design** 体系：

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI 组件库**: **Ant Design (AntD) 5.x** —— 提供专业的金融级 UI 组件。
- **图表库**: ECharts (echarts-for-react) —— 负责多维度资产可视化。
- **路由**: React Router v6 (支持 ProtectedRoute 路由守卫)。
- **通信**: Axios —— 预留与 Go (Gin) 后端的 RESTful API 对接。

## 📂 规范化目录结构

根据项目需求，当前 `src` 目录结构如下：

```text
src/
├── api/                    # Axios 请求封装与后端接口定义 (对接人: 顾晨旻)
├── components/
│   ├── layout/            # 全局布局 (Layout.tsx 整合了侧边栏与导航)
│   └── ProtectedRoute.tsx # 登录权限校验
├── pages/                 # 业务页面
│   ├── Login.tsx          # 登录页 (含 Token 模拟存储)
│   ├── Dashboard.tsx      # 投资驾驶舱 (资产统计 + ECharts 概览)
│   ├── Upload.tsx         # 数据导入 (支持 AntD Dragger 拖拽上传)
│   ├── Analysis.tsx       # AI 风险分析 (时间轴模式识别)
│   ├── Prediction.tsx     # 趋势预测 (AI 模拟曲线)
│   └── History.tsx        # 历史归档 (数据表格)
├── router/                # 路由配置中心
├── App.tsx                # 根组件 (配置 ConfigProvider 中文化)
└── main.tsx               # 入口文件 (引入 antd/dist/reset.css)
```

## 📅 当前功能进度 (2026年3月)

### 已完成 ✅
- **UI 框架迁移**: 完成从 Tailwind 到 Ant Design 的全面重构，界面更专业。
- **路由系统**: 实现了完善的嵌套路由与登录拦截逻辑。
- **登录模块**: 具备基础的登录状态管理 (localStorage)。
- **页面骨架**: 全功能页面 (Dashboard, Upload, Analysis...) 已完成 AntD 静态化开发。

### 待完成 / 进行中 🚧
- **后端联调**: 对接 Go 后端的 `/api/v1` 接口，实现真实数据流。
- **数据预览**: 在 `Upload` 页面增加上传后的数据预校验表格。
- **AI 动态加载**: 优化 `Analysis` 页面的加载状态，模拟 AI 生成报告过程。
- **PDF 导出**: 增加 `html2canvas` + `jspdf` 实现分析报告一键导出。

## 🛠 开发环境启动

1. **安装依赖**：
   ```bash
   npm install
   ```
2. **启动开发服务器**：
   ```bash
   npm run dev
   ```
3. **配置文件说明**：
   - 现已弃用 `tailwind.config.js`，样式由 AntD 全权接管。
   - API 地址请在 `src/api/index.ts` 中统一修改。

---
**武大计算机学院 · 计算机综合项目实践 · 小组作业**
```

---