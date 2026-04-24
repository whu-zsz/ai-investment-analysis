-- 创建数据库
CREATE DATABASE IF NOT EXISTS stock_analysis
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE stock_analysis;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar_url VARCHAR(500),
    investment_preference VARCHAR(20) DEFAULT 'balanced',
    total_profit DECIMAL(15,2) DEFAULT 0.00,
    risk_tolerance VARCHAR(10) DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 交易记录表
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    transaction_date DATE NOT NULL,
    transaction_type VARCHAR(20) NOT NULL COMMENT 'buy, sell, dividend',
    asset_type VARCHAR(20) NOT NULL COMMENT 'stock, fund, bond',
    asset_code VARCHAR(20) NOT NULL,
    asset_name VARCHAR(100) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    price_per_unit DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    commission DECIMAL(10,2) DEFAULT 0.00,
    profit DECIMAL(15,2),
    notes TEXT,
    source_file VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_date (user_id, transaction_date),
    INDEX idx_asset_code (asset_code),
    INDEX idx_transaction_type (transaction_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易记录表';

-- 持仓明细表
CREATE TABLE IF NOT EXISTS portfolios (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    asset_code VARCHAR(20) NOT NULL,
    asset_name VARCHAR(100) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    total_quantity DECIMAL(10,2) NOT NULL,
    available_quantity DECIMAL(10,2) NOT NULL,
    average_cost DECIMAL(10,2) NOT NULL,
    current_price DECIMAL(10,2),
    market_value DECIMAL(15,2),
    profit_loss DECIMAL(15,2),
    profit_loss_percent DECIMAL(5,2),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_user_asset (user_id, asset_code),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='持仓明细表';

-- AI分析报告表
CREATE TABLE IF NOT EXISTS ai_analysis_reports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    report_type VARCHAR(20) NOT NULL COMMENT 'summary, risk, prediction, pattern',
    report_title VARCHAR(200) NOT NULL,
    analysis_period_start DATE NOT NULL,
    analysis_period_end DATE NOT NULL,
    total_investment DECIMAL(15,2) NOT NULL,
    total_profit DECIMAL(15,2) NOT NULL,
    profit_rate DECIMAL(10,4) NOT NULL,
    risk_level VARCHAR(10) NOT NULL COMMENT 'low, medium, high',
    investment_style VARCHAR(50),
    summary_text TEXT NOT NULL,
    risk_analysis TEXT,
    pattern_insights TEXT,
    prediction_text TEXT,
    chart_data JSON,
    recommendations TEXT,
    ai_model VARCHAR(50) DEFAULT 'deepseek',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_type (user_id, report_type),
    INDEX idx_period (analysis_period_start, analysis_period_end),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI分析报告表';

-- 上传文件记录表
CREATE TABLE IF NOT EXISTS uploaded_files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(10) NOT NULL COMMENT 'csv, xlsx, xls',
    upload_status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending, processing, success, failed',
    records_imported INT DEFAULT 0,
    error_message TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    INDEX idx_user_uploaded (user_id, uploaded_at),
    INDEX idx_upload_status (upload_status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上传文件记录表';

-- 插入测试用户（密码：123456）
INSERT INTO users (username, email, password_hash, investment_preference, total_profit, risk_tolerance)
VALUES
('testuser', 'test@example.com', '$2a$10$sNiBl7K8mOVa8PQyWKFTKuM.rryXJ1ZAGuMyU9/AA9fYxs9W04guW', 'balanced', 0.00, 'medium');

-- 完成
SELECT 'Database initialization completed!' AS message;
