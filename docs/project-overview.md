# 项目全景文档

**项目名称**: demo (认证演示服务)
**最后更新**: 2026-03-27
**基于代码实际存在内容生成，不包含猜测内容**

---

## 1. 项目整体架构

### 技术栈
- **编程语言**: Go 1.26.1
- **框架**: Kratos v2.8.4 (微服务框架)
- **ORM**: GORM v1.25.12
- **数据库**: MySQL 8 (通过 `gorm.io/driver/mysql`)
- **缓存**: Redis 7.2.5 (已配置但当前业务未使用)
- **API 定义**: Protobuf (支持 HTTP + gRPC 双协议)
- **依赖注入**: Google Wire

### 分层架构 (DDD 风格，遵循严格单向依赖)

```
api/            → Protobuf API 定义和生成代码
cmd/server/     → 应用入口 (main)
configs/        → 配置文件
internal/
├── server/     → 传输层 (HTTP/gRPC 服务器启动) ← 相当于 Controller
├── service/    → 服务层 (API 适配层，入参出参转换)
├── biz/        → 业务用例层 (业务流程编排)
├── repo/       → 仓储层 (领域实体定义 + Repository 接口定义)
├── data/       → 数据访问层 (Repository 实现，DAO)
└── conf/       → 配置结构定义
migrations/     → 数据库迁移 SQL 脚本
docs/          → 项目文档
```

### 各层职责边界

| 层级 | 位置 | 职责 | 依赖方向 |
|------|------|------|----------|
| **Transport** | `internal/server` | 启动 HTTP/gRPC 服务器，注册服务 | `server → service` |
| **Service** | `internal/service` | Protobuf DTO 转换，调用 Biz 层用例 | `service → biz` |
| **Biz** | `internal/biz` | 业务流程编排、核心业务逻辑 | `biz → repo` |
| **Repo** | `internal/repo` | 领域实体定义 + Repository 接口定义 | 无对外依赖 |
| **Data** | `internal/data` | Repository 实现，GORM 数据访问 | `data → repo` (实现接口) |

**依赖规则** (严格遵守):
- 单向依赖: `server → service → biz → repo ← data`
- 禁止反向依赖
- 禁止跨层直接调用

---

## 2. 核心数据模型及其关系

### 数据模型分布

所有领域实体定义在 `internal/repo/` 目录:

| 实体 | 文件 | 说明 | 对应数据库表 |
|------|------|------|-------------|
| `Greeter` | `repo/greeter.go` | 示例实体 | 无（示例未实际建表）|
| `User` | `repo/user.go` | 用户 | `users` |
| `UserLoginLog` | `repo/user_login_log.go` | 用户登录记录 | `tbl_user_login_logs_00` ~ `tbl_user_login_logs_09` (10张分表) |

### 实体定义

#### User (用户)
```go
type User struct {
    ID        int64         // 用户ID (主键)
    Username  string        // 用户名 (唯一)
    Email     string        // 邮箱 (唯一)
    Password  string        // 哈希后的密码
    CreatedAt time.Time     // 创建时间
    UpdatedAt time.Time     // 更新时间
}
```

#### UserLoginLog (用户登录记录)
```go
type UserLoginLog struct {
    ID          int64         // 登录记录ID (主键)
    UserID      int64         // 关联用户ID
    LoginIP     string        // 登录IP地址
    UserAgent   string        // 用户代理 (浏览器/设备信息)
    LoginStatus int16         // 登录状态 (0-失败, 1-成功)
    FailReason  string        // 失败原因
    LoginTime   time.Time     // 登录时间
    GeoLocation *string       // 地理位置 (可选)
    TokenID     *string       // JWT 令牌ID (可选)
    CreatedAt   time.Time     // 创建时间
    UpdatedAt   time.Time     // 更新时间
}
```

### 分表设计 (UserLoginLog)

- **分表策略**: 按 `userID % 10` 哈希取模水平分表
- **表数量**: 10 张 (`tbl_user_login_logs_00` ~ `tbl_user_login_logs_09`)
- **分表逻辑**: GORM PO 的 `TableName()` 方法根据 UserID 动态计算表名
- **查询路由**:
  - Create / FindByUserID: userID 已知，直接路由到对应分表
  - FindByTokenID: tokenID 不含 userID，需要遍历所有 10 张分表

### 数据库表结构

#### users 表 (001_create_users_table.sql)
```sql
CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) DEFAULT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY idx_username (username),
    UNIQUE KEY idx_email (email),
    KEY idx_status (status)
);
```

#### tbl_user_login_logs_xx 分表 (002_create_user_login_logs_tables.sql × 10)
```sql
CREATE TABLE tbl_user_login_logs_xx (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    login_ip VARCHAR(50) NOT NULL DEFAULT '',
    user_agent VARCHAR(500) NOT NULL DEFAULT '',
    login_status SMALLINT NOT NULL DEFAULT 0 COMMENT '0-失败, 1-成功',
    fail_reason VARCHAR(255) NOT NULL DEFAULT '',
    login_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    geo_location VARCHAR(100) DEFAULT NULL,
    token_id VARCHAR(100) DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (id),
    KEY idx_tbl_user_login_logs_xx_user_id (user_id),
    KEY idx_tbl_user_login_logs_xx_login_time (login_time),
    KEY idx_tbl_user_login_logs_xx_login_ip (login_ip),
    KEY idx_tbl_user_login_logs_xx_user_id_login_status (user_id, login_status)
);
```

### ER 关系

```
┌─────────┐         ┌─────────────────┐
│  User  │    1:N  │ UserLoginLog    │
└─────────┘         └─────────────────┘
  (id)            (user_id → id)
```

---

## 3. 关键业务流程

### 用户注册流程

```
HTTP/gRPC Request
    ↓
service/AuthService.Register
    - 从 Protobuf DTO 提取参数
    ↓
biz/AuthUseCase.Register
    1. 检查用户名是否已存在  → repo.UserRepo.FindByUsername
    2. 检查邮箱是否已存在    → repo.UserRepo.FindByEmail
    3. bcrypt 加密密码
    4. 创建用户            → repo.UserRepo.Create
    5. 生成 JWT 令牌
    6. 返回用户信息和令牌
    ↓
service/AuthService
    - 转换为 Protobuf 响应
    ↓
HTTP/gRPC Response
```

**调用链路**: `server → service.AuthService → biz.AuthUseCase → repo.UserRepo ← data.userRepo (GORM)`

### 用户登录流程 (含登录日志记录)

```
HTTP/gRPC Request (username/email + password)
    ↓
service/AuthService.Login
    - 提取 clientIP (当前留空，需从 HTTP context 获取)
    - 提取 userAgent (当前留空，需从 HTTP context 获取)
    ↓
biz/AuthUseCase.Login
    1. 根据用户名/邮箱查找用户 → repo.UserRepo.FindByUsername/FindByEmail
    2. 验证密码 (bcrypt.CompareHashAndPassword)
    3. 生成 JWT 令牌
    4. **记录登录日志** → repo.UserLoginLogRepo.Create
       - 无论成功失败都会记录
       - 包含 IP、UserAgent、状态、失败原因
    5. 返回用户信息和令牌
    ↓
service/AuthService
    - 转换为 Protobuf 响应
    ↓
HTTP/gRPC Response
```

**调用链路**: `server → service → biz.AuthUseCase → [repo.UserRepo, repo.UserLoginLogRepo] ← data`

### 动态分表写入示例

当用户ID = 123 登录成功:
1. `tableIndex = 123 % 10 = 3`
2. 写入 `tbl_user_login_logs_03`
3. GORM 根据 `userLoginLogPO.TableName()` 自动路由

---

## 4. 外部依赖和集成点

### 直接依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/go-kratos/kratos/v2` | v2.8.4 | 微服务框架 (transport、middleware、config、log 等) |
| `gorm.io/gorm` | v1.25.12 | ORM 框架 |
| `gorm.io/driver/mysql` | v1.5.7 | MySQL 驱动 |
| `github.com/golang-jwt/jwt/v5` | v5.1.0 | JWT 令牌生成与验证 |
| `golang.org/x/crypto` | v0.48.0 | bcrypt 密码加密 |
| `github.com/google/wire` | v0.6.0 | 编译期依赖注入 |
| `google.golang.org/grpc` | v1.68.1 | gRPC 支持 |
| `google.golang.org/protobuf` | v1.36.1 | Protobuf 编码 |

### 外部系统集成

| 系统 | 用途 | 状态 |
|------|------|------|
| **MySQL** | 主数据存储 | 已集成 |
| **Redis** | 缓存 | 已配置连接，但当前业务代码未使用 |
| **无** | 消息队列 | 未集成 |
| **无** | 第三方 API 调用 | 无 |

### 配置

配置文件: `configs/config.yaml`

```yaml
server:
  http:
    addr: 0.0.0.0:8000
  grpc:
    addr: 0.0.0.0:9000
data:
  database:
    driver: mysql
    source: root:12345678@tcp(127.0.0.1:3306)/test?parseTime=True
  redis:
    addr: 127.0.0.1:6379
auth:
  jwt_secret: "your-secret-key-change-in-production"
```

---

## 5. 应用入口服务差异

**实际情况**: 项目只包含**一个单体服务**，没有多个入口。

| 入口 | 位置 | 职责 |
|------|------|------|
| **主服务** | `cmd/server/main.go` | 同时启动 HTTP 服务器 (8000 端口) 和 gRPC 服务器 (9000 端口)，包含所有业务功能 |

> 用户问题中提到的 `main/teacher-api`、`main/teacher-admin`、`main/teacher-job` 在当前代码库中**不存在**。

---

## 6. 现有测试覆盖情况

**当前状态**: 项目中**不存在任何测试文件**。

- 未发现任何 `*_test.go` 文件
- Makefile 中有 `test` 和 `test-coverage` 目标，但没有实际测试代码
- 测试覆盖率: 0%

### 可用的 Make 测试目标

```make
make test           # 运行所有测试 (当前无测试)
make test-coverage  # 生成覆盖率报告 (当前无测试)
```

---

## 7. 项目统计

| 指标 | 数值 |
|------|------|
| Go 源文件 (.go) | ~ 25 个 |
-  API 生成代码: 8
-  internal: 15
-  cmd: 3
| 数据库表 | 11 张 |
| users: 1 |
| tbl_user_login_logs: 10 (分表) |
| 单元测试 | 0 个 |
| 直接依赖 | 8 个 |

---

## 8. 开发命令参考 (来自 Makefile)

```bash
# 安装开发工具 (protoc 插件、wire、golangci-lint)
make install-tools

# 根据 proto 文件生成 Go 代码
make proto

# 生成 Wire 依赖注入代码
make wire

# 本地编译
make build

# 运行
make run

# 代码格式化
make fmt

# 静态代码检查
make lint

# 运行测试
make test
```

---

## 待人工确认

- [ ] Redis 缓存计划用于哪些场景（当前已配置连接但未使用）
- [ ] 是否计划添加单元测试
- [ ] 地理位置信息 (geo_location) 是否需要通过 IP 库自动解析
- [ ] clientIP / UserAgent 获取需要在 service 层完善实现（当前留空）

---

**本文档由 Claude 自动生成，基于代码库实际存在内容，不包含猜测**
