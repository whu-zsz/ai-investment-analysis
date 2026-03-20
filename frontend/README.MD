
# AI投顾助手 · 前端部分

基于 React + TypeScript + Vite + Tailwind CSS 构建的投资记录分析与预测 Web 应用前端。

## 项目概述

- 目标：为普通投资者提供上传投资记录 → AI 自动分析 → 可视化报告 → 趋势预测 的完整 Web 体验
- 技术栈：
  - React 18 + TypeScript
  - Vite（构建工具）
  - Tailwind CSS（样式）
  - react-router-dom v6（路由）
  - react-hook-form（表单）
  - echarts-for-react（图表，计划使用）
  - axios（API 请求，计划使用）
  - JWT + localStorage（模拟登录与路由保护）

## 当前功能状态（2026年3月）

已完成：
- 整体布局（左侧 Sidebar + 顶部 Navbar + 主内容区）
- 登录页面（Login.tsx）+ 简单模拟登录验证
- 路由保护（ProtectedRoute）
- 首页 Dashboard（静态概览卡片）
- 上传页面（Upload.tsx，静态拖拽区域占位）
- 分析报告页面（Analysis.tsx，静态卡片占位）
- 趋势预测页面（Prediction.tsx，静态预测卡片占位）
- 历史记录页面（History.tsx，静态表格占位）
- 响应式基础适配（Tailwind sm/md/lg）

待完成 / 进行中：
- 真实拖拽上传 + 文件预览 + 上传进度条（react-dropzone）
- Analysis 页面完整实现（Tabs + ECharts 图表：折线图、饼图、热力图）
- Prediction 页面图表叠加（历史 + 预测曲线）
- 前后端 API 联调（axios 封装 + 类型定义）
- PDF 报告一键导出（jspdf 或调用后端接口）
- 移动端底部导航 / 抽屉菜单优化
- 错误处理、Loading 状态、骨架屏
- 暗黑模式（可选）

## 项目结构
```bash
frontend/
├── public/                     # 静态资源
│   └── favicon.ico             # （建议替换成项目图标）
│
├── src/
│   ├── components/             # 可复用组件
│   │   ├── layout/             # 全局布局相关
│   │   │   ├── Layout.tsx
│   │   │   ├── Navbar.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── ProtectedRoute.tsx   # 路由守卫
│   │   └── ui/                 # 通用 UI 组件（计划放置）
│   │       └── UploadZone.tsx  # （拖拽上传组件，待完善）
│   │
│   ├── pages/                  # 页面级组件（路由对应）
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Upload.tsx
│   │   ├── Analysis.tsx
│   │   ├── Prediction.tsx
│   │   └── History.tsx
│   │
│   ├── router/                 # 路由配置
│   │   └── index.tsx
│   │
│   ├── types/                  # TypeScript 类型定义（计划放置 API 类型）
│   │   └── index.ts
│   │
│   ├── services/               # API 请求封装（待创建）
│   │   └── api.ts              # （计划）
│   │
│   ├── App.tsx                 # 根组件（渲染路由）
│   ├── index.css               # Tailwind 全局样式入口
│   └── main.tsx                # React 渲染入口
│
├── .gitignore
├── eslint.config.js            # （可选，ESLint 配置）
├── index.html                  # Vite 入口 HTML
├── package.json
├── package-lock.json
├── postcss.config.js           # Tailwind PostCSS 配置
├── README.md                   # 本文件
├── tailwind.config.js          # Tailwind 配置
├── tsconfig.app.json
├── tsconfig.json
├── tsconfig.node.json
└── vite.config.ts              # Vite 配置

## 快速启动

# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
# 默认打开 http://localhost:5173

测试账号（当前模拟登录）
用户名：盛哲
密码：123456