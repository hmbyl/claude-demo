# 用户登录注册数据表创建计划

## Context

需要在本机 MySQL 的 `test` 数据库中创建用户登录注册相关的数据表，为后续实现用户认证功能做准备。

## 方案概述

创建一个 `users` 表，包含用户注册和登录所需的核心字段，同时考虑安全性和可扩展性。

## 数据表设计

### users 表结构

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT UNSIGNED | 主键 | AUTO_INCREMENT, PRIMARY KEY |
| username | VARCHAR(50) | 用户名 | UNIQUE, NOT NULL |
| email | VARCHAR(100) | 邮箱 | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | 密码哈希 | NOT NULL |
| nickname | VARCHAR(50) | 昵称 | NULLABLE |
| avatar | VARCHAR(255) | 头像URL | NULLABLE |
| status | TINYINT | 状态: 0-禁用, 1-正常 | DEFAULT 1 |
| created_at | DATETIME | 创建时间 | NOT NULL, DEFAULT CURRENT_TIMESTAMP |
| updated_at | DATETIME | 更新时间 | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP |

### 索引设计
- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_username` (`username`)
- UNIQUE KEY `idx_email` (`email`)
- INDEX `idx_status` (`status`)

## 实施步骤

1. 创建 SQL 迁移文件 `migrations/001_create_users_table.sql`
2. 执行 SQL 在 test 数据库中创建表

## SQL 文件内容

```sql
-- 创建 users 表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    email VARCHAR(100) NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    nickname VARCHAR(50) DEFAULT NULL COMMENT '昵称',
    avatar VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-禁用, 1-正常',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY idx_username (username),
    UNIQUE KEY idx_email (email),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

## 注意事项

- 密码只存储哈希值，不存储明文
- 使用 utf8mb4 字符集支持完整的 Unicode（包括 emoji）
- username 和 email 都有唯一索引，保证唯一性
