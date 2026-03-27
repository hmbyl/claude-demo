# SKILL: generate-unit-test
## 触发命令: /gen-test {文件路径}
## 执行步骤:
1. 分析目标文件的所有公开函数和方法
2. 为每个函数生成 Table-Driven 测试用例，覆盖：
    - 正常路径（至少 2 个用例）
    - 边界条件（空值、零值、最大值）
    - 错误路径（参数非法、依赖失败）
3. 使用 testify/assert 断言
4. Mock 外部依赖（使用 mockery 生成的 mock）
5. 运行测试确认全部通过
6. 输出覆盖率报告
