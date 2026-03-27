# 用户登录验证码校验 - 执行计划

## 任务分解（自底向上，按DDD分层）

---

### Task 1: 新增Go依赖（go-redis）
- **文件**: `go.mod`
- **依赖**: 无
- **验收**: `go mod tidy` 成功，无冲突
- **预估**: 0.25h

---

### Task 2: 新增验证码配置结构定义
- **文件**: `internal/conf/conf.go`
- **依赖**: Task 1 (依赖已完成，只是新增代码，无代码依赖)
- **验收**:
  - 增加 `Captcha` 结构体，包含所有配置字段
  - 在 `Bootstrap` 结构体中增加 `Captcha *Captcha` 字段
  - 字段标签 `yaml` 正确，可正确反序列化
- **预估**: 0.25h

---

### Task 3: 定义CaptchaRepo接口（Domain层）
- **文件**: `internal/repo/captcha.go`
- **依赖**: Task 2（无代码依赖，只是接口定义，可并行）
- **验收**:
  - 定义 `CaptchaRepo` 接口，包含所有需要的方法
  - 方法签名符合业务需求：`StoreCaptcha`, `GetCaptcha`, `DeleteCaptcha`, `GetFailureCount`, `IncrementFailureCount`, `ResetFailureCount`
  - 接口定义在domain/repo层，符合DDD原则
- **预估**: 0.5h

---

### Task 4: 初始化Redis客户端（Data层基础设施）
- **文件**: `internal/data/data.go`
- **依赖**: Task 1 (go-redis依赖), Task 2 (conf有Redis配置)
- **验收**:
  - 在 `Data` 结构体中增加 `rdb *redis.Client` 字段
  - 在 `NewData` 函数中，根据配置初始化Redis客户端
  - 清理函数中关闭Redis客户端
  - 可正确连接Redis，连接失败返回错误
- **预估**: 0.5h

---

### Task 5: 实现CaptchaRepo Redis版本（Data层）
- **文件**: `internal/data/captcha_repo.go`
- **依赖**: Task 3 (接口定义), Task 4 (Redis客户端初始化完成)
- **验收**:
  - 定义 `captchaRepo` 结构体，持有 `*Data` 引用
  - 实现 `CaptchaRepo` 接口的所有方法
  - Redis Key命名符合设计规范
  - 过期时间设置正确
  - `IncrementFailureCount` 在首次自增时正确设置过期时间
- **预估**: 1h

---

### Task 6: 修改AuthUseCase增加验证码业务逻辑（Biz层）
- **文件**: `internal/biz/auth_usecase.go`
- **依赖**: Task 2 (配置), Task 3 (CaptchaRepo接口)
- **验收**:
  - 在 `AuthUseCase` 结构体中增加 `captchaRepo CaptchaRepo` 和 `captchaConfig *conf.Captcha` 字段
  - 在 `NewAuthUseCase` 构造函数中注入依赖
  - 增加方法：`isIPWhitelisted(ip string) bool` - IP白名单判断，启动时预解析CIDR
  - 增加方法：`needCaptcha(ctx context.Context, clientIP string, loginBy string) bool` - 判断是否需要验证码
  - 增加方法：`verifyCaptcha(ctx context.Context, captchaId, code string) (bool, error)` - 验证码校验
  - 增加方法：`generateCaptcha(ctx context.Context) (string, string, error)` - 生成验证码
  - 修改 `Login` 方法，插入验证码校验流程，遵循设计文档中的流程
  - 登录成功后正确清零IP和用户名两个维度的失败计数
  - 登录失败正确增加失败计数
  - 验证码错误时失败计数不累加（符合需求）
  - 验证码生成满足：4位，必须同时包含数字和字母
- **预估**: 2.5h

---

### Task 7: 修改API proto定义（API层）
- **文件**: `api/auth/v1/auth.proto`
- **依赖**: Task 3（无代码依赖，接口定义，可并行）
- **验收**:
  - 新增 `GetCaptchaRequest` 和 `GetCaptchaReply` 消息定义
  - 在 `Auth` service中新增 `GetCaptcha` rpc方法
  - 在 `LoginRequest` 中新增可选字段 `captcha_id` 和 `captcha_code`
  - proto格式正确，可正常编译
- **预估**: 0.5h

---

### Task 8: 修改AuthService实现新接口（Service层）
- **文件**: `internal/service/auth.go`
- **依赖**: Task 6 (Biz层AuthUseCase修改完成), Task 7 (proto编译完成，生成go代码)
- **验收**:
  - 修改 `Login` 方法，从请求中获取 `captcha_id` 和 `captcha_code`，传递给 `uc.Login`
  - 实现 `GetCaptcha` 方法，调用 `uc.GenerateCaptcha`，返回正确响应
  - 错误处理正确，返回对应错误信息
  - 从context中正确提取clientIP（框架层面已经支持，需要确保传递给UseCase）
- **预估**: 1h

---

### Task 9: 重新生成proto go代码
- **文件**: `api/auth/v1/*.pb.go` （自动生成）
- **依赖**: Task 7 (proto修改完成)
- **验收**: 生成代码成功，无编译错误
- **预估**: 0.25h

---

### Task 10: 修改wire依赖注入配置
- **文件**: 需要检查哪个文件提供ProviderSet
- **依赖**: Task 5 (CaptchaRepo实现完成), Task 6 (AuthUseCase修改需要新依赖)
- **验收**:
  - 在 `ProviderSet` 中注册 `NewCaptchaRepo`
  - `NewAuthUseCase` 的依赖注入正确包含 `CaptchaRepo` 和 `CaptchaConfig`
  - `wire` 编译成功
- **预估**: 0.5h

---

### Task 11: 添加配置文件示例
- **文件**: `configs/config.yaml`（或相应环境配置）
- **依赖**: Task 2 (配置结构定义完成)
- **验收**:
  - 增加 `captcha` 配置段落
  - 包含所有配置项：`enabled`, `threshold`, `window_minutes`, `expire_minutes`, `ip_whitelist`
  - 配置示例正确
- **预估**: 0.25h

---

### Task 12: 编写单元测试
- **文件**: `internal/biz/auth_usecase_test.go`, `internal/data/captcha_repo_test.go`
- **依赖**: Task 6 (biz完成), Task 5 (data完成)
- **验收**:
  - 覆盖验证码生成逻辑
  - 覆盖是否需要验证码判断逻辑
  - 覆盖IP白名单匹配逻辑（包含CIDR测试）
  - 覆盖验证码校验逻辑
  - 覆盖失败计数增减逻辑
  - 覆盖边界条件：阈值刚好等于、刚好过期、验证码错误等
  - 测试通过率100%
- **预估**: 2h

---

### Task 13: 代码编译检查
- **文件**: 整个项目
- **依赖**: 所有上述任务完成
- **验收**: `go build ./...` 成功，无编译错误
- **预估**: 0.25h

---

## 任务总计
- 总任务数: 13个
- 预估总工作量: **~10小时**

---

## 开发顺序与关键路径

### 关键路径（必须顺序执行）

```
Task 1 (add deps)
    ↓
Task 2 (config struct)
    ↓
Task 4 (init redis client)
    ↓
Task 3 (repo interface) → 可以和Task 2并行
Task 7 (proto api) → 可以和Task 3并行
    ↓
Task 5 (captcha repo impl)
    ↓
Task 9 (generate proto code) ← 依赖Task 7
    ↓
Task 6 (biz useCase)
    ↓
Task 8 (service layer) ← 依赖Task 6和Task 9
    ↓
Task 10 (wire inject)
    ↓
Task 11 (config example)
    ↓
Task 12 (unit tests)
    ↓
Task 13 (compile check)
```

### 可并行任务
- **Task 3** (repo接口定义) 可以在 **Task 2** 完成后和 **Task 4** 并行
- **Task 7** (proto API定义) 可以在 **Task 2** 完成后和 **Task 3/4** 并行

### 甘特图示意（按天）

| 时间段 | 任务 |
|--------|------|
| 第1小时 | Task 1, Task 2 |
| 第2小时 | Task 3, Task 4, Task 7 (并行) |
| 第3小时 | Task 5, Task 9 |
| 第4-6小时 | Task 6 |
| 第7小时 | Task 8 |
| 第8小时 | Task 10, Task 11 |
| 第9-11小时 | Task 12 (单元测试) |
| 第12小时 | Task 13 收尾 |

---

## 数据库变更清单

**✓ 本需求不需要数据库变更**

- 不需要新增表
- 不需要修改现有表结构
- 不需要migration脚本
- 原有 `tbl_user_login_logs` 保持不变，继续用于登录日志审计

---

## 测试策略

### 分层测试策略

| 模块 | 测试类型 | 覆盖目标 | 测试要点 |
|------|----------|----------|----------|
| **CaptchaRepo (Redis实现)** | 单元测试（可用miniredis mock） | 100% 覆盖所有方法 | - 存储/获取/删除验证码正确<br>- 过期时间设置正确<br>- 失败计数自增正确<br>- 重置计数正确 |
| **IP白名单匹配** | 单元测试 | 100% 覆盖 | - 单个IP匹配正确<br>- CIDR网段匹配正确<br>- 不匹配正确返回<br>- 空白名单正确处理 |
| **验证码生成** | 单元测试 | 100% 覆盖 | - 生成固定4位长度<br>- 必须同时包含数字和字母<br>- 随机性满足要求 |
| **AuthUseCase 业务流程** | 单元测试（mock CaptchaRepo） | 核心流程100%覆盖 | - 功能关闭跳过验证码 ✓<br>- IP白名单跳过验证码 ✓<br>- 失败次数未达到阈值跳过 ✓<br>- 达到阈值需要验证码 ✓<br>- 需要验证码但未传 → 错误 ✓<br>- 验证码过期 → 错误 ✓<br>- 验证码错误 → 错误，失败计数不累加 ✓<br>- 验证码正确 → 继续校验密码 ✓<br>- 密码错误 → 失败计数+1 ✓<br>- 密码正确 → 清零失败计数 ✓<br>- IP或任一维度达到阈值即触发 ✓ |
| **API接口** | 集成测试 | 覆盖接口调用 | - GetCaptcha可正常返回 ✓<br>- Login带验证码可正常校验 ✓ |

### 测试数据构造要点

| 测试场景 | 测试数据构造 |
|----------|--------------|
| CIDR网段匹配 | 构造 `192.168.0.0/16`，测试 `192.168.1.1` (匹配), `192.169.1.1` (不匹配) |
| 刚好达到阈值 | 构造失败次数 = 阈值 → 应该触发 |
| 刚好低于阈值 | 构造失败次数 = 阈值-1 → 不触发 |
| 验证码过期 | 存储验证码后手动修改过期时间 → 应该返回过期错误 |
| 验证码大小写 | 存储 `Ab12`，输入 `ab12` → 应该通过（不区分大小写） |

---

## 风险清单及应对

| 风险 | 影响概率 | 影响程度 | 应对措施 |
|------|----------|----------|----------|
| **Redis连接失败服务启动失败** | 低 | 高 | Redis连接失败在 `NewData` 阶段返回错误，服务启动失败，便于发现配置问题 |
| **Redis不可用导致登录完全不可用** | 低 | 高 | 设计方案：如果功能关闭，跳过Redis操作；如果功能开启Redis不可用，返回错误明确提示；也可以考虑降级：当Redis不可用时暂时跳过验证码校验 |
| **GetCaptcha被恶意调用耗尽Redis内存** | 中 | 中 | 应对：一期上线先观察，二期增加基于IP的限流，当前设计预留了扩展能力 |
| **客户端不传递验证码导致登录失败** | 低（新接口前端会配合修改） | 中 | 应对：错误消息明确提示"需要验证码"，前端可以根据错误提示引导用户获取验证码 |
| **白名单配置语法错误导致服务启动失败** | 低 | 中 | 应对：启动时预解析，任何CIDR解析错误直接返回错误，提前暴露配置问题，不带到运行时 |
| **go-redis版本冲突** | 低 | 低 | 选用和现有kratos框架兼容的版本，`github.com/go-redis/redis/v8` 是稳定版本 |
| **同一IP并发登录失败导致计数不准确** | 低 | 低 | Redis `Incr` 是原子操作，计数不会出错，天然支持并发 |

---

## 验收标准

整体功能验收满足以下所有条件：

1. ✓ 编译成功：`go build ./...` 无错误
2. ✓ 单元测试：所有测试通过
3. ✓ 功能开启时，连续失败达到阈值后强制验证码
4. ✓ 任一维度（IP/用户名）达到阈值即触发
5. ✓ 验证码存储Redis，5分钟过期
6. ✓ 验证码校验不区分大小写
7. ✓ 验证码错误不增加失败计数
8. ✓ 登录成功清零失败计数
9. ✓ IP白名单支持单个IP和CIDR网段
10. ✓ 配置全部外部化，支持修改阈值、窗口时间、过期时间
11. ✓ 遵循项目编码规范（DDD分层，依赖方向正确）
