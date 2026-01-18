# 迁移到 Query 包指南

本文档说明如何从旧的 `model` 包迁移到新的 `query` 包。

## 核心变化

### 1. ServiceContext 变化

**旧代码**:
```go
type ServiceContext struct {
    Conn                     sqlx.SqlConn
    UsersModel               model.UsersModel
    WebauthnCredentialsModel model.WebauthnCredentialsModel
    UserAccessLogsModel      model.UserAccessLogsModel
}
```

**新代码**:
```go
type ServiceContext struct {
    DB                  *sql.DB
    Users               *query.UsersCustom
    WebauthnCredentials *query.WebauthnCredentialsCustom
    AuditLogs           *query.AuditLogsCustom
}
```

### 2. 查询方法变化

#### FindOne -> FindByPK 或自定义方法

**旧代码**:
```go
user, err := svcCtx.UsersModel.FindOne(ctx, uid)
user, err := svcCtx.UsersModel.FindOneByUsername(ctx, username)
```

**新代码**:
```go
// 主键查询
user, err := svcCtx.Users.FindByPK(ctx, uid)

// 自定义查询（需要在 users.go 中添加）
user, err := svcCtx.Users.FindByUsername(ctx, username)
```

#### Insert -> Create

**旧代码**:
```go
result, err := svcCtx.UserAccessLogsModel.Insert(ctx, &model.UserAccessLogs{
    UserId: uid,
    Action: "LOGIN",
})
```

**新代码**:
```go
result, err := svcCtx.AuditLogs.Create(ctx, &query.AuditLogs{
    UserID: uid,
    Action: "LOGIN",
})
```

#### Update -> UpdateByPK 或 Update

**旧代码**:
```go
err := svcCtx.UsersModel.Update(ctx, user)
```

**新代码**:
```go
// 方式1: 更新整个对象
err := svcCtx.Users.UpdateByPK(ctx, user)

// 方式2: 更新指定字段
_, err := svcCtx.Users.
    Where(query.usersField.ID.Eq(user.ID)).
    Update(ctx, map[string]interface{}{
        "github_id": githubID,
    })
```

### 3. 添加自定义方法

在 `internal/query/{table}.go` 中添加业务逻辑：

```go
// users.go
package query

import "context"

type UsersCustom struct {
    *usersDo
}

func NewUsers(db Executor) *UsersCustom {
    return &UsersCustom{
        usersDo: users.WithDB(db).(*usersDo),
    }
}

// FindByUsername 根据用户名查询用户
func (c *UsersCustom) FindByUsername(ctx context.Context, username string) (*Users, error) {
    users, err := c.Where(usersField.Username.Eq(username)).Find(ctx)
    if err != nil {
        return nil, err
    }
    if len(users) == 0 {
        return nil, ErrRecordNotFound
    }
    return users[0], nil
}

// FindByOAuth 根据第三方账号查询用户
func (c *UsersCustom) FindByOAuth(ctx context.Context, provider string, oauthID string) (*Users, error) {
    var cond WhereCondition
    switch provider {
    case "github":
        cond = usersField.GithubID.Eq(oauthID)
    case "google":
        cond = usersField.GoogleID.Eq(oauthID)
    default:
        return nil, ErrRecordNotFound
    }
    
    users, err := c.Where(cond).Find(ctx)
    if err != nil {
        return nil, err
    }
    if len(users) == 0 {
        return nil, ErrRecordNotFound
    }
    return users[0], nil
}
```

## 迁移步骤

### 1. ServiceContext (✅ 已完成)
- [x] 更改导入
- [x] 更新字段类型
- [x] 初始化 Query 访问器
- [x] 添加 DB.Close()

### 2. Middleware
- [ ] AuthMiddleware - 更新 UserAccessLogsModel
- [ ] MFAMiddleware - 如有使用
- [ ] MFAStepUpMiddleware - 如有使用

### 3. Logic 层
- [ ] auth_login_logic.go
- [ ] auth_oauth2_callback_logic.go
- [ ] 其他 logic 文件

### 4. 添加自定义方法
- [ ] users.go - FindByUsername, FindByOAuth
- [ ] audit_logs.go - Insert 方法（如需要）
- [ ] webauthn_credentials.go - 相关查询

## 类型映射

| 旧类型 (model) | 新类型 (query) |
|---------------|---------------|
| `model.Users` | `query.Users` |
| `model.UserAccessLogs` | `query.AuditLogs` |
| `model.WebauthnCredentials` | `query.WebauthnCredentials` |
| `sqlx.ErrNotFound` | `query.ErrRecordNotFound` |

## 注意事项

1. **字段名变化**: 数据库列名转为 Go 字段名
   - `user_id` -> `UserID`
   - `github_id` -> `GithubID`

2. **NULL 字段**: 使用指针或 sql.Null* 类型
   - `sql.NullString` 保持不变
   - 或使用 `*string` (需要检查生成的类型)

3. **事务**: 使用 `*sql.Tx` 替代 `sqlx.Session`
   ```go
   tx, _ := svcCtx.DB.BeginTx(ctx, nil)
   defer tx.Rollback()
   
   usersInTx := query.NewUsers(tx)
   _, err := usersInTx.Create(ctx, user)
   
   tx.Commit()
   ```

4. **错误处理**: 统一使用 `query.ErrRecordNotFound`
   ```go
   if err == query.ErrRecordNotFound {
       // 处理未找到
   }
   ```

## 优势

✅ **完全类型安全** - 所有查询在编译期检查
✅ **链式调用** - 更优雅的查询语法
✅ **自动补全** - IDE 完美支持
✅ **索引优化** - FindByIndex 高性能查询
✅ **自定义方法** - 安全添加业务逻辑，永不被覆盖

## 示例对比

### 查询用户

**旧方式**:
```go
user, err := svcCtx.UsersModel.FindOneByUsername(ctx, "alice")
if err == sqlx.ErrNotFound {
    return nil, errors.New("user not found")
}
```

**新方式**:
```go
user, err := svcCtx.Users.FindByUsername(ctx, "alice")
if err == query.ErrRecordNotFound {
    return nil, errors.New("user not found")
}
```

### 复杂查询

**新方式的优势**:
```go
// 类型安全的复杂查询
users, err := svcCtx.Users.
    Where(query.usersField.IsActive.IsTrue()).
    Where(query.usersField.CreatedAt.Gte(time.Now().Add(-24*time.Hour))).
    Order(query.usersField.CreatedAt.Desc()).
    Limit(10).
    Find(ctx)
```

### 更新字段

**旧方式**: 需要查询、修改、更新
```go
user, _ := svcCtx.UsersModel.FindOne(ctx, uid)
user.GithubId = sql.NullString{String: githubID, Valid: true}
err := svcCtx.UsersModel.Update(ctx, user)
```

**新方式**: 直接更新
```go
_, err := svcCtx.Users.
    Where(query.usersField.ID.Eq(uid)).
    Update(ctx, map[string]interface{}{
        "github_id": githubID,
    })
```

## 获取帮助

查看详细文档：
- `/Users/x/dev/work/gpmg/README.md` - GPMG 使用指南
- `/Users/x/dev/work/gpmg/CUSTOM_METHODS.md` - 自定义方法指南
- `/Users/x/dev/work/alpha-trade/internal/query/CUSTOM_README.md` - 项目自定义方法示例
