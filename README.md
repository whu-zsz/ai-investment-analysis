# AI投顾助手 - 基于大模型的投资记录分析与预测系统

武汉大学计算机学院 · 计算机综合项目实践课程设计  
指导教师：谭小琼  
小组成员：张盛哲（前端）、顾晨旻（后端）、林润民（测试 & 文档）

## 项目简介

本项目是一个面向普通投资者的 Web 应用，用户可上传个人投资记录（CSV/Excel），通过调用大语言模型 API 实现：

- 投资行为总结与偏好分析
- 盈亏统计与风险评估
- 投资模式识别与预警
- 未来趋势预测与可视化
- 一键生成 PDF 分析报告

目标用户：大学生、白领、股市散户等非专业投资者。

## 技术栈

- 前端：React 18 + TypeScript + Vite + Tailwind CSS + react-router-dom v6 + ECharts
- 后端：Go + Gin + GORM + MySQL 8.0
- AI 集成：DeepSeek / OpenAI GPT-4o-mini / 通义千问 等大模型 API
- 可视化：ECharts
- 认证：JWT
- 部署计划：阿里云/腾讯云学生服务器

## 目录结构
AI-INVESTMENT/
├── backend/          # Go 后端服务
├── frontend/         # React + Vite 前端
└── README.md
