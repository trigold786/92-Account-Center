# Stage 1.0 — 项目全面审查与优化设计

## 概述

对 92-Account-Center 项目进行全面审查，修复关键 Bug，优化代码架构和质量，更新过期文档，添加缺失注释，完成 Stage 1.0 交付。

## 审计发现摘要

| 类别 | 数量 | 严重级别 |
|------|------|---------|
| 关键 Bug（错误端口/URL） | 2 | CRITICAL |
| 文档过期（端口/架构） | 4 | HIGH |
| 代码质量（错误处理/DRY） | 5 | HIGH/MEDIUM |
| 前端质量（类型/console.log） | 3 | MEDIUM/LOW |
| 配置不一致 | 2 | MEDIUM |

## 修复计划

### Phase A: 关键 Bug 修复

1. **payment-service/cmd/main.go:231** — CREDIT_SERVICE_URL 默认值从 30317 改为 30312
2. **account-service/cmd/main.go:114** — AUTH_SERVICE_URL 默认值从 30300 改为 30302

### Phase B: 代码架构优化

3. **提取 getEnv 到 pkg/env/env.go** — 消除 9 个服务中的重复代码
4. **修复高风险静默错误** — 至少处理 io.ReadAll、RowsAffected、json.Marshal 的错误
5. **fmt.Printf → 结构化日志** — pkg/saga/orchestrator.go、pkg/async/publisher.go
6. **panic → error return** — account-service/pkg/crypto/key_manager.go
7. **CORS 源可配置** — api-gateway/internal/middleware/cors.go 使用环境变量

### Phase C: 文档更新

8. **ARCHITECTURE.md** — 更新端口号和架构图
9. **API_SPEC.md** — 更新所有端口号
10. **DEPLOYMENT.md** — 更新端口号，移除 Kafka 引用
11. **.env.example** — 添加 20+ 缺失的环境变量
12. **CI/CD** — 修复 Go 版本号

### Phase D: 前端优化

13. **移除 console.log** — 生产代码中的调试输出
14. **添加 TypeScript 类型** — 减少 `any` 使用
15. **RBAC 路由守卫** — admin/finance 路由权限检查

### Phase E: 注释与规范

16. **Godoc 注释** — 为导出的类型和函数添加文档注释
17. **.gitattributes** — 统一行尾符

## 约束

- 不改变业务逻辑
- 不删除现有功能
- 所有修改必须通过构建和测试
- 保持向后兼容

## 验证

- `go build ./...` 每个服务
- `go test ./...` 每个服务
- `npm run build` web-ui
- `golangci-lint run` 无新增警告
