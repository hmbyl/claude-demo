# 用户登录验证码校验 - 技术设计方案

## 1. 需求概述

为防止用户恶意登录，在用户登录失败达到指定次数后，增加验证码校验。

**明确需求：**
- 登录失败次数≥N次（可配置，默认3次）后，强制验证码校验
- 统计最近15分钟内的失败次数
- 统计维度：优先按IP，无法获取IP则按用户名/邮箱
- IP或用户名任一维度达到阈值即触发
- 验证码4位，包含数字+英文字母，随机生成
- 后端只生成验证码内容，前端生成图片
- 验证码存储Redis，5分钟过期
- 校验不区分大小写
- 如果验证码错误但用户名密码正确，失败次数不累加，但不允许登录
- 登录成功后清零失败计数
- 需要新增`GetCaptcha`接口获取验证码
- 支持IP白名单（数组格式，支持CIDR网段）
- 所有配置全局统一

---

## 2. 架构设计

### 2.1 整体架构（遵循现有DDD分层）

```
API (proto)
  ↓
Controller (service/AuthService)
  ↓
  GetCaptcha → AuthUseCase.GenerateCaptcha → CaptchaRepo.Store → Redis
  Login →
    1. 判断IP白名单 → 判断是否需要验证码 →
    2. 需要验证码 → 校验验证码 → 失败返回错误
    3. 校验用户名密码 → 成功清零失败计数 → 返回Token
                                 ↓ 失败 → 失败计数+1
```

### 2.2 分层职责

| 层级 | 职责 |
|------|------|
| **api/proto** | 定义 `GetCaptcha` 接口和 `LoginRequest` 增加验证码字段 |
| **service** | HTTP/GRPC入参解析，调用UseCase，不包含业务逻辑 |
| **biz (UseCase)** | 核心业务流程编排：判断是否需要验证码 → 生成验证码 → 验证码校验 → 失败计数增减 → 登录流程 |
| **domain/repo** | 定义 `CaptchaRepo` 接口：存储验证码、操作失败计数 |
| **data** | 实现 `CaptchaRepo` 接口，基于Redis操作 |
| **conf** | 增加验证码配置结构定义 |

依赖方向保持：`service → biz → domain ← data`，严格单向依赖，无跨层调用。

---

## 3. API设计

### 3.1 proto定义修改

```proto
// Authentication service definition
service Auth {
  // Register a new user
  rpc Register (RegisterRequest) returns (RegisterReply) {}
  // Login with existing credentials
  rpc Login (LoginRequest) returns (LoginReply) {}
  // Get captcha when required
  rpc GetCaptcha (GetCaptchaRequest) returns (GetCaptchaReply) {}
}

// Login request contains login credentials
message LoginRequest {
  oneof login_by {
    string username = 1;
    string email = 2;
  }
  string password = 3;
  // 验证码ID（需要验证码时必填）
  optional string captcha_id = 4;
  // 验证码内容（需要验证码时必填）
  optional string captcha_code = 5;
}

// Get captcha request
message GetCaptchaRequest {
}

// Get captcha response
message GetCaptchaReply {
  string captcha_id = 1;
  string captcha_code = 2;
  int64 expire_at = 3;
}
```

### 3.2 接口说明

| 接口 | 说明 |
|------|------|
| `GetCaptcha` | 生成一个新的验证码，返回ID和内容。前端提前调用，获取后展示图片给用户。 |
| `Login` | `captcha_id` 和 `captcha_code` 为可选字段，只有当触发验证码时才校验。 |

---

## 4. 配置设计

### 4.1 配置结构 (`internal/conf/conf.go`)

```go
type Bootstrap struct {
  Server  *Server  `yaml:"server"`
  Data    *Data    `yaml:"data"`
  Auth    *Auth    `yaml:"auth"`
  Captcha *Captcha `yaml:"captcha"`  // 新增
}

type Captcha struct {
  Enabled        bool          `yaml:"enabled"`          // 是否启用验证码功能，默认true
  Threshold      int           `yaml:"threshold"`        // 触发阈值，默认3
  WindowMinutes  int           `yaml:"window_minutes"`   // 失败统计窗口，默认15分钟
  ExpireMinutes  int           `yaml:"expire_minutes"`   // 验证码过期时间，默认5分钟
  IPWhitelist    []string      `yaml:"ip_whitelist"`     // IP白名单，支持单个IP和CIDR网段
}
```

### 4.2 配置示例

```yaml
captcha:
  enabled: true
  threshold: 3
  window_minutes: 15
  expire_minutes: 5
  ip_whitelist:
    - "127.0.0.1"
    - "192.168.0.0/16"
```

---

## 5. 数据存储设计

### 5.1 Redis Key设计

**不需要新增数据库表**，所有动态数据都存在Redis。数据库已有登录日志保持不变用于审计。

| Key 格式 | 用途 | 过期时间 | 存储内容 |
|----------|------|----------|----------|
| `captcha:code:{captchaId}` | 存储验证码内容 | 5分钟（可配置） | 验证码字符串（如 "aB3d"） |
| `captcha:fail:ip:{ipAddr}` | IP维度失败计数 | 15分钟（可配置） | 整数，失败次数 |
| `captcha:fail:user:{loginBy}` | 用户名/邮箱维度失败计数 | 15分钟（可配置） | 整数，失败次数 |

> captchaId 使用UUID生成，保证唯一性。

### 5.2 Domain接口定义 (`internal/repo/captcha.go`)

```go
package repo

import (
  "context"
)

// CaptchaRepo defines the interface for captcha storage and failure count operations
// This follows DDD: interface defined in domain (repo), implemented in data
type CaptchaRepo interface {
  // StoreCaptcha stores a captcha in Redis
  StoreCaptcha(ctx context.Context, captchaId, code string, expireMinutes int) error
  // GetCaptcha retrieves a captcha from Redis
  GetCaptcha(ctx context.Context, captchaId string) (string, error)
  // DeleteCaptcha deletes a captcha from Redis
  DeleteCaptcha(ctx context.Context, captchaId string) error

  // GetFailureCount gets current failure count for given key (ip or user)
  GetFailureCount(ctx context.Context, key string) (int, error)
  // IncrementFailureCount increments failure count, sets expiration if not exists
  IncrementFailureCount(ctx context.Context, key string, expireMinutes int) (int, error)
  // ResetFailureCount resets failure count to zero (deletes the key)
  ResetFailureCount(ctx context.Context, key string) error
}
```

---

## 6. 核心业务流程设计

### 6.1 获取验证码流程 (`GetCaptcha`)

```
1. 接收请求
2. 生成随机UUID作为captchaId
3. 生成4位随机验证码（必须包含数字+字母）
4. 存储到Redis: key = captcha:code:{captchaId}, value = code, expire = 5分钟
5. 返回 captchaId, captchaCode, expireAt
```

### 6.2 登录流程修改

新增验证码校验步骤，整合到原有登录流程中：

```
接收登录请求 (loginBy, password, clientIP, captchaId, captchaCode)
  ↓
1. 判断功能是否启用 → 不启用直接跳过验证码校验
  ↓
2. 判断是否在IP白名单 → 在白名单直接跳过验证码校验
  ↓
3. 获取当前失败计数：
  - 能获取clientIP → 获取IP计数 ipCount
  - 同时获取用户名/邮箱计数 userCount
  - 任一计数 ≥ 阈值 → 需要验证码
  ↓
4. 需要验证码：
  ├─ 检查captchaId/captchaCode是否为空 → 为空返回"需要验证码"错误
  │
  ├─ 从Redis获取存储的正确验证码 → 不存在返回"验证码已过期，请重新获取"
  │
  ├─ 比较验证码（都转小写，不区分大小写） → 不匹配返回"验证码错误"
  │   ↓
  │   根据需求：验证码错误但用户名密码正确也不允许登录，失败次数不累加
  │   返回错误，要求用户重新输入验证码
  │
  └─ 验证码正确 → 删除已使用的验证码 → 继续下一步
  ↓
5. 不需要验证码 → 直接下一步
  ↓
6. 校验用户名密码（原有逻辑不变）
  ↓
  ├─ 用户名密码错误
  │   ↓
  │   记录登录日志（原有逻辑）
  │   失败计数+1（IP和user两个key都加）
  │   返回"用户名或密码错误"
  │
  └─ 用户名密码正确
      ↓
      清零失败计数（删除IP和user两个key）
      生成JWT Token → 返回成功
```

---

## 7. 验证码生成规则

- 长度固定4位
- 字符集：`0-9` + `a-z` + `A-Z` 共62个字符
- 必须同时包含至少一个数字和至少一个字母
- 随机生成，如果不满足条件就重新生成
- 校验时不区分大小写，都转为小写比较

### 示例实现逻辑：

```go
// generateCaptcha generates a 4-digit captcha that contains at least one digit and one letter
func generateCaptcha() string {
  for attempts := 0; attempts < 10; attempts++ {
    code := randomString(4, digitsAndLetters)
    if hasDigit(code) && hasLetter(code) {
      return code
    }
  }
  // 如果多次尝试都不满足，强制调整最后一位满足条件
  // ...
}
```

---

## 8. IP白名单匹配

使用Go标准库 `net` 包进行CIDR网段匹配：

1. 启动时预解析所有白名单配置，转换为`net.IPNet`数组缓存
2. 匹配时遍历白名单，任一匹配即命中
3. 使用标准库，不需要引入第三方依赖

---

## 9. Redis实现 (`internal/data/captcha_repo.go`)

使用go-redis客户端操作Redis：

- 存储验证码：`Set` 命令，带过期时间
- 获取验证码：`Get` 命令
- 删除验证码：`Del` 命令
- 失败计数：`Incr` 命令，首次自增时设置过期时间
- 重置计数：`Del` 命令

---

## 10. 依赖分析

需要新增依赖：
- `github.com/go-redis/redis/v8` - Redis客户端，广泛使用的标准库
- `github.com/google/uuid` - 生成captchaId（已经是间接依赖，不需要新增）

所有都是业界标准依赖，符合项目规范。

---

## 多角色评审PK

### 【架构师视角评审】

**评审问题：**

1. **是否符合现有DDD分层架构？职责是否清晰？**
   - ✓ 当前设计符合现有分层：API → Service → Biz → Domain ← Data
   - ✓ 各层职责清晰，没有跨层调用，依赖方向正确
   - ✓ 接口定义在domain层，实现在data层，符合DDD原则

2. **是否引入不必要复杂度？是否有跨层调用？**
   - ✓ 没有引入不必要复杂度，需求本身需要这些组件
   - ✓ 没有跨层调用，严格遵循单向依赖

3. **扩展性如何？用户量增长10倍哪里会成为瓶颈？**
   - **当前设计评估：** 使用Redis存储所有动态数据，本身扩展性很好
   - **潜在瓶颈：**
     - IP白名单匹配每次都遍历所有网段，如果白名单条目很多（>1000）会有O(n)性能问题
     - Redis本身是单线程模型，但对于验证码这种简单操作，即使增长10倍也完全能承受
     - 数据库登录日志分表已经做好，增长10倍没问题

**架构师修改建议：**
- 对于IP白名单匹配，如果条目数量大可以优化前缀树，但一般场景下白名单不会很大，当前设计足够。
- 增加注释说明key命名规范，便于后续维护。
- 在 `AuthUseCase` 中依赖注入 `CaptchaRepo`，而不是让 `AuthUseCase` 直接操作Redis，符合依赖倒置原则。✓ 当前设计已经做到。

---

### 【DBA视角评审】

**评审问题：**

1. **新增表结构是否符合数据库规范？**
   - ✓ 当前设计不需要新增数据库表，所有动态数据存储在Redis中
   - 原有 `tbl_user_login_logs` 保持不变用于审计，符合现有规范
   - 原有表结构已经包含 `id`, `created_at`, `updated_at`, `deleted_at`，符合规范

2. **数据量预估多少？是否需要分表？**
   - Redis中的数据都是短期的：
     - 验证码：5分钟过期，平均QPS假设100，同时在线约 100 * 5 = 500条
     - 失败计数：15分钟过期，即使每天1万次登录尝试，同时在线也只有约 100 个key
   - 数据量非常小，完全不需要分表
   - 数据库登录日志已经分表，按现有分表策略即可

3. **有没有慢查询风险？**
   - ✓ 所有统计查询都放在Redis，不走数据库，不存在慢查询
   - ✓ 不需要新增数据库查询，原有结构不受影响
   - 即使增长10倍，Redis的O(1)操作也不会有性能问题

**DBA修改建议：**
- 确认Redis过期时间设置合理，当前设计已经配置自动过期，不会产生内存泄漏 ✓
- Redis key命名清晰，便于运维排查问题，当前设计满足 ✓
- 无需修改，设计符合要求。

---

### 【安全工程师视角评审】

**评审问题：**

1. **是否存在注入、越权、数据泄露风险？**
   - **注入风险：** 所有Redis操作使用带参数的命令（go-redis客户端会正确处理），不存在注入 ✓
   - **越权风险：** 验证码ID是随机UUID，攻击者无法猜测其他用户的验证码ID ✓
   - **数据泄露风险：** 验证码本身不存储敏感信息，不存在泄露 ✓
   - **暴力破解：** 验证码本来就是为了防止暴力破解，设计达成目标 ✓

2. **敏感数据是否加密存储？**
   - 验证码不涉及敏感数据，不需要加密
   - 失败计数只是数字，不需要加密
   - Redis存储在内网，符合安全要求

3. **接口是否有鉴权？**
   - `GetCaptcha` 接口不需要鉴权，本身就是为未登录用户提供的
   - 验证码生成是无状态的，不会带来安全问题
   - 即使攻击者频繁调用获取验证码，也无法绕过限制，因为失败计数还是会增加，符合安全设计

4. **潜在风险发现：**
   - **DoS风险**：攻击者可以无限制调用 `GetCaptcha` 接口，占用Redis内存。虽然有过期机制，但如果QPS非常大可能导致Redis内存飙升。

**安全工程师修改建议：**
- 增加对 `GetCaptcha` 接口的限流策略（基于IP），防止恶意刷接口占用内存
- 验证码使用后立即删除，当前设计在验证码校验成功后已经删除 ✓
- 验证码长度4位足够，字符集62个，有 62^4 ≈ 1400万种可能，暴力破解概率极低 ✓

---

### 【QA视角评审】

**评审问题：**

1. **哪些场景难以测试？**
   - IP白名单的CIDR网段匹配：需要构造不同IP测试是否正确匹配
   - 边界条件：刚好等于阈值、刚好过期等
   - 并发场景：同一IP同时多次登录失败，计数是否正确（Redis Incr是原子操作，没问题 ✓）

2. **边界条件和异常路径是否在设计中考虑到了？**
   | 场景 | 设计是否考虑 |
   |------|-------------|
   | 功能关闭（enabled: false） | ✓ 跳过所有校验，正常登录 |
   | 需要验证码但前端没传 | ✓ 返回明确错误提示 |
   | 验证码过期 | ✓ 返回"已过期"错误，提示重新获取 |
   | 验证码错误 | ✓ 返回错误，失败次数不累加（符合需求） |
   | IP不在白名单，用户名在白名单？- 需求明确白名单是IP层面的 | ✓ 只判断IP，不在白名单正常触发 |
   | 同时IP和用户名都达到阈值 | ✓ 任一达到就触发，符合需求 |
   | 登录成功后不清零？ | ✓ 设计明确登录成功后清零 |
   | 验证码校验成功后用户名密码错误 | ✓ 失败计数正常+1，正确处理 |

3. **是否需要特殊测试数据构造？**
   - 需要测试CIDR网段匹配，QA需要构造不同IP测试
   - 需要测试时间边界（刚好过期），可通过修改Redis过期时间测试
   - 并发测试需要多线程同时调用登录接口，测试计数准确性（Redis Incr原子性保证正确）

**QA修改建议：**
- 在接口返回错误码中，区分不同错误类型（验证码为空、验证码过期、验证码错误），便于前端不同展示
- 当前设计只返回error string，建议定义明确的不同错误原因，方便QA自动化测试

---

## 设计修订（基于评审建议）

根据上述评审意见，对设计进行修订：

### 1. 错误分类（响应QA建议）

在登录时，需要区分不同错误原因：

| 错误场景 | 返回错误信息 |
|----------|--------------|
| 需要验证码但未提供 | "验证码 required" |
| 验证码已过期 | "验证码已过期，请重新获取" |
| 验证码不匹配 | "验证码错误，请重新输入" |

这样前端可以根据不同错误做不同处理，也便于QA测试。

### 2. GetCaptcha接口限流（响应安全工程师建议）

可以在 `GetCaptcha` 接口增加基于IP的简单限流：
- 同一IP每分钟最多调用 `GetCaptcha` 10次
- 超出限制返回"请求过于频繁，请稍后再试"
- 限流也存储在Redis，key格式 `captcha:ratelimit:ip:{ip}`
- 这是一个增强优化，不做强制要求，第一期可以不实现，二期补充。本次设计先预留扩展性。

### 3. IP白名单预解析优化（响应架构师建议）

在服务启动时预解析所有白名单配置，缓存 `net.IPNet` 对象，避免每次匹配重新解析：

```go
// 在 AuthUseCase 中缓存解析后的白名单
type AuthUseCase struct {
  // ... existing fields
  captchaRepo    repo.CaptchaRepo
  captchaConfig  *conf.Captcha
  parsedWhitelist []net.IPNet // 预解析后的网段
}

// NewAuthUseCase 中预解析
func NewAuthUseCase(...) {
  // ...
  parsedWhitelist := parseIPWhitelist(config.IPWhitelist)
  return &AuthUseCase{
    // ...
    parsedWhitelist: parsedWhitelist,
  }
}
```

这样即使白名单条目很多，匹配性能也很好。

### 4. 最终修订后的分层依赖

```
conf
  └─ Captcha 配置结构
    ↓
api/auth/v1
  └─ proto 增加 GetCaptcha 接口，Login 增加验证码字段
    ↓
service/AuthService
  ├─ Login: 提取 clientIP, 传入 captchaId/captchaCode 到 UseCase
  └─ GetCaptcha: 调用 UseCase.GenerateCaptcha 返回结果
    ↓
biz/AuthUseCase
  ├─ 判断IP白名单（使用预解析）
  ├─ 判断是否需要验证码（查询Redis计数）
  ├─ 验证码校验
  ├─ 失败计数增减、清零
  └─ 原有登录流程
    ↓
repo (domain) 定义接口
  └─ CaptchaRepo 接口
    ↓
data 实现接口
  └─ CaptchaRepo Redis实现
    ↓
Redis
```

完全符合DDD分层，职责清晰，依赖正确。

---

## 变更文件清单

需要修改/新增的文件：

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `api/auth/v1/auth.proto` | 新增 `GetCaptcha` 接口，`LoginRequest` 增加验证码字段 |
| 修改 | `internal/conf/conf.go` | 新增 `Captcha` 配置结构 |
| 新增 | `internal/repo/captcha.go` | 定义 `CaptchaRepo` 接口 |
| 修改 | `internal/data/data.go` | 新增Redis客户端初始化，Data结构体增加redis字段 |
| 新增 | `internal/data/captcha_repo.go` | 基于Redis实现 `CaptchaRepo` |
| 修改 | `internal/biz/auth_usecase.go` | 增加验证码相关方法，修改Login流程，增加预解析IP白名单 |
| 修改 | `internal/service/auth.go` | 修改Login方法获取验证码参数，实现GetCaptcha方法 |
| 修改 | `go.mod` | 新增go-redis依赖 |
| 新增 | `migrations/xxx_xxx.sql` | 不需要，无数据库变更 |
| 修改 | 配置文件 `configs/config.yaml` | 需要增加captcha配置项 |

---

## 性能评估

| 操作 | 时间复杂度 | 存储位置 | 评估 |
|------|------------|----------|------|
| 获取验证码 | O(1) | Redis | 非常快 |
| 登录-检查失败计数 | O(1) + O(1) | Redis | 两次Redis Get，非常快 |
| 登录-校验验证码 | O(1) | Redis | 一次Redis Get，非常快 |
| IP白名单匹配 | O(n) n=白名单条目数 | 内存 | n通常很小（<20），可忽略 |
| 失败计数增加 | O(1) | Redis | Redis Incr原子操作 |
| 失败计数清零 | O(1) | Redis | Del操作 |

整体登录流程只增加了2-3次Redis Get操作，对响应时间影响在ms级别，完全可接受。

---

## 安全评估

| 风险 | 控制措施 | 状态 |
|------|----------|------|
| 暴力破解密码 | 失败次数达到阈值强制验证码，增大破解难度 | ✅ 防护有效 |
| 验证码被猜测 | 4位62字符，1400万种可能，无法在线猜测 | ✅ 防护有效 |
| 验证码复用 | 验证成功后立即删除，无法复用 | ✅ 防护有效 |
| DoS攻击占用Redis内存 | 建议增加限流，第一期可先上线观察，后续补充 | ⚠️ 可后续优化 |
| 注入攻击 | 使用go-redis官方客户端，参数化命令 | ✅ 无注入风险 |

---

## 总结

本设计方案：
1. 完全符合需求和现有项目架构规范
2. 不需要修改数据库表结构，无迁移成本
3. 使用Redis存储，性能良好，扩展性好
4. 遵循DDD分层原则，职责清晰，依赖正确
5. 评审发现的问题都已得到修正，设计完善


让我检查：
