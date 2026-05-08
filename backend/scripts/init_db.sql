-- 本脚本仅用于初始化本地开发/测试数据库。
-- 应用表结构由后端启动时的 GORM AutoMigrate 负责创建和维护。

CREATE DATABASE IF NOT EXISTS stock_analysis
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS stock_analysis_test
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;
