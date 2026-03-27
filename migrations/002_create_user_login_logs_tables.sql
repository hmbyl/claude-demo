-- 创建用户登录日志分表（共10张表，按 user_id 哈希取模分表）
-- 分表算法: table_index = user_id % 10，表名格式: tbl_user_login_logs_xx (xx 为 00-09)

-- 创建分表 00
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_00 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_00_user_id (user_id),
    KEY idx_tbl_user_login_logs_00_login_time (login_time),
    KEY idx_tbl_user_login_logs_00_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_00_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_00';

-- 创建分表 01
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_01 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_01_user_id (user_id),
    KEY idx_tbl_user_login_logs_01_login_time (login_time),
    KEY idx_tbl_user_login_logs_01_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_01_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_01';

-- 创建分表 02
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_02 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_02_user_id (user_id),
    KEY idx_tbl_user_login_logs_02_login_time (login_time),
    KEY idx_tbl_user_login_logs_02_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_02_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_02';

-- 创建分表 03
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_03 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_03_user_id (user_id),
    KEY idx_tbl_user_login_logs_03_login_time (login_time),
    KEY idx_tbl_user_login_logs_03_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_03_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_03';

-- 创建分表 04
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_04 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_04_user_id (user_id),
    KEY idx_tbl_user_login_logs_04_login_time (login_time),
    KEY idx_tbl_user_login_logs_04_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_04_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_04';

-- 创建分表 05
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_05 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_05_user_id (user_id),
    KEY idx_tbl_user_login_logs_05_login_time (login_time),
    KEY idx_tbl_user_login_logs_05_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_05_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_05';

-- 创建分表 06
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_06 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_06_user_id (user_id),
    KEY idx_tbl_user_login_logs_06_login_time (login_time),
    KEY idx_tbl_user_login_logs_06_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_06_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_06';

-- 创建分表 07
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_07 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_07_user_id (user_id),
    KEY idx_tbl_user_login_logs_07_login_time (login_time),
    KEY idx_tbl_user_login_logs_07_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_07_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_07';

-- 创建分表 08
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_08 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_08_user_id (user_id),
    KEY idx_tbl_user_login_logs_08_login_time (login_time),
    KEY idx_tbl_user_login_logs_08_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_08_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_08';

-- 创建分表 09
CREATE TABLE IF NOT EXISTS tbl_user_login_logs_09 (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '登录记录ID',
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '登录IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户代理（浏览器/设备信息）',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '登录状态: 0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    geo_location VARCHAR(100) DEFAULT NULL COMMENT '地理位置',
    token_id VARCHAR(100) DEFAULT NULL COMMENT '令牌ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_09_user_id (user_id),
    KEY idx_tbl_user_login_logs_09_login_time (login_time),
    KEY idx_tbl_user_login_logs_09_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_09_user_id_login_status (user_id, login_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志分表_09';
