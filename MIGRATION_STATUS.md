# Query 包迁移状态

## ✅ 已完成

### 1. ServiceContext 迁移 (100%)
- [x] 更改导入从 `model` 到 `query`
- [x] 更改数据库连接从 `sqlx.SqlConn` 到 `*sql.DB`
- [x] 初始化 Query 访问器
  - `Users *query.UsersCustom`
  - `WebauthnCredentials *query.WebauthnCredentialsCustom`
  - `AuditLogs *query.AuditLogsCustom`
- [x] 添加 `DB.Close()` 到 `ServiceContext.Close()`

**文件**: `/Users/x/dev/work/alpha-trade/internal/svc/service_context.go`

### 2. Query 自定义方法 (100%)

#### users.go
- [x] `FindByUsername(ctx, username)` - 根据用户名查询
- [x] `FindByOAuth(ctx, provider, oauthID)` - OAuth 登录查询
- [x] `GetRevokedAt(ctx, userID)` - 获取撤销时间

#### audit_logs.go  
- [x] `RecordAction(ctx, userID, ip, action...)` - 记录审计日志

#### webauthn_credentials.go
- [x] 基础结构

**文件**: 
- `/Users/x/dev/work/alpha-trade/internal/query/users.go`
- `/Users/x/dev/work/alpha-trade/internal/query/audit_logs.go`
- `/Users/x/dev/work/alpha-trade/internal/query/webauthn_credentials.go`

### 3. Query 包验证
- [x] 所有生成的代码编译通过
- [x] 自定义方法编译通过
- [x] 类型安全验证

## ⏳ 待完成

### 1. Middleware 迁移 (0/3)

需要迁移的文件：
- [ ] `internal/middleware/auth_middleware.go`
  - 更改 `model.UserAccessLogsModel` 到 `*query.AuditLogsCustom`
  - 更新 `Insert` 调用到 `RecordAction`
  
- [ ] `internal/middleware/mfa_middleware.go`
  - 检查是否使用数据库
  
- [ ] `internal/middleware/mfa_step_up_middleware.go`
  - 检查是否使用数据库

### 2. Logic 层迁移 (0/11)

需要迁移的文件：
- [ ] `internal/logic/auth/auth_login_logic.go`
  - `UsersModel.FindOneByUsername` → `Users.FindByUsername`
  - `UserAccessLogsModel.Insert` → `AuditLogs.RecordAction`
  
- [ ] `internal/logic/auth/auth_o_auth2_callback_logic.go`
  - `UsersModel.FindOne` → `Users.FindByPK`
  - `UsersModel.FindOneByOAuth` → `Users.FindByOAuth`
  - `UsersModel.Update` → `Users.UpdateByPK` 或 `Where().Update()`
  - `UserAccessLogsModel.Insert` → `AuditLogs.RecordAction`
  
- [ ] `internal/logic/auth/auth_logout_logic.go`
- [ ] `internal/logic/auth/auth_o_auth2_init_logic.go`
- [ ] `internal/logic/system/system_info_logic.go`
- [ ] `internal/logic/auth/passkey/*.go` (6 files)

### 3. Revocation Manager 接口适配

**文件**: `/Users/x/dev/work/alpha-trade/internal/pkg/revocation/revocation.go`

需要确保 `RevocationManager` 可以使用新的 `*query.UsersCustom`：
```go
// 旧接口可能期望
type RevocationManager interface {
    IsRevoked(ctx context.Context, userID int64, issuedAt time.Time) bool
}

// 需要确认内部是否使用 model.UsersModel
```

## 📋 迁移对照表

| 旧代码 (model) | 新代码 (query) |
|---------------|---------------|
| `svcCtx.UsersModel` | `svcCtx.Users` |
| `svcCtx.UserAccessLogsModel` | `svcCtx.AuditLogs` |
| `svcCtx.WebauthnCredentialsModel` | `svcCtx.WebauthnCredentials` |
| `model.Users` | `query.Users` |
| `model.UserAccessLogs` | `query.AuditLogs` |
| `FindOne(ctx, id)` | `FindByPK(ctx, id)` |
| `FindOneByUsername(ctx, username)` | `FindByUsername(ctx, username)` |
| `FindOneByOAuth(ctx, provider, id)` | `FindByOAuth(ctx, provider, id)` |
| `Insert(ctx, model)` | `Create(ctx, model)` |
| `Update(ctx, model)` | `UpdateByPK(ctx, model)` |
| `sqlx.ErrNotFound` | `query.ErrRecordNotFound` |

## 🔧 迁移命令

### 批量替换导入
```bash
cd /Users/x/dev/work/alpha-trade

# 替换 import
find internal/logic internal/middleware -name "*.go" -exec sed -i '' 's|github.com/iluyuns/alpha-trade/internal/model|github.com/iluyuns/alpha-trade/internal/query|g' {} +

# 替换类型
find internal/logic internal/middleware -name "*.go" -exec sed -i '' 's|model\.Users|query.Users|g' {} +
find internal/logic internal/middleware -name "*.go" -exec sed -i '' 's|model\.UserAccessLogs|query.AuditLogs|g' {} +
```

### 验证编译
```bash
cd /Users/x/dev/work/alpha-trade
go build ./internal/...
```

## 📖 参考文档

- `/Users/x/dev/work/alpha-trade/MIGRATION_TO_QUERY.md` - 完整迁移指南
- `/Users/x/dev/work/gpmg/README.md` - GPMG 使用文档
- `/Users/x/dev/work/gpmg/CUSTOM_METHODS.md` - 自定义方法指南

## ⚠️ 注意事项

1. **字段名变化**: 
   - `user_id` → `UserID`
   - `github_id` → `GithubID` (string 类型，不是 sql.NullString)
   - `google_id` → `GoogleID` (string 类型，不是 sql.NullString)

2. **AuditLogs 表结构变化**:
   - 原 `user_access_logs` 表改为 `audit_logs`
   - 字段: `UserID`, `IpAddress`, `Action`, `TargetType`, `TargetID`, `Changes`, `IsVerified`
   - 不再有: `UserAgent`, `Status`, `Reason`, `Details`

3. **事务使用**:
   ```go
   tx, _ := svcCtx.DB.BeginTx(ctx, nil)
   defer tx.Rollback()
   
   usersInTx := query.NewUsers(tx)
   auditInTx := query.NewAuditLogs(tx)
   
   // 执行操作...
   
   tx.Commit()
   ```

## 🚀 下一步

1. 运行自动替换命令（谨慎）
2. 手动迁移 middleware 文件（3 个）
3. 手动迁移 logic 文件（11 个）
4. 更新 RevocationManager
5. 编译验证
6. 运行测试

## 进度统计

- **总文件数**: 17 个
- **已完成**: 4 个 (ServiceContext + 3个自定义文件)
- **待完成**: 13 个
- **完成度**: 24%

---

**最后更新**: 2026-01-18 22:19
**状态**: 🟡 进行中
