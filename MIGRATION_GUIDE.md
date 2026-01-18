# 从 gozero-pg-model-gen 迁移到 gpmg

本指南帮助你将 alpha-trade 项目从 gozero-pg-model-gen 迁移到 gpmg。

## 主要变化

| 方面 | gozero-pg-model-gen | gpmg |
|-----|-------------------|------|
| 包路径 | `internal/model` | `internal/query` |
| 命名风格 | `UserModel` | `user` (全局变量) |
| DO 结构体 | 无 | `userDo` |
| 接口 | `UserModel` 接口 | `IUserDo` 接口 |
| Field 访问 | `UserFields.Age` | `userField.Age` |
| 方法风格 | `SelectBuilder()` | `WithDB().Where().Find()` |

## 代码迁移对照表

### 1. 初始化

#### 旧代码 (gozero-pg-model-gen)
```go
import "alpha-trade/internal/model"

userModel := model.NewUserModel(conn)
```

#### 新代码 (gpmg)
```go
import "alpha-trade/internal/query"

userDO := query.user.WithDB(db)
```

### 2. 创建记录

#### 旧代码
```go
user := &model.User{
    Username: "alice",
    Email:    "alice@example.com",
}
result, err := userModel.Insert(ctx, user)
// 或者
insertedUser, err := userModel.InsertReturn(ctx, session, user)
```

#### 新代码
```go
user := &query.User{
    Username: "alice",
    Email:    "alice@example.com",
}
insertedUser, err := userDO.Create(ctx, user)
```

### 3. 根据主键查询

#### 旧代码
```go
user, err := userModel.FindOne(ctx, userID)
if err == model.ErrNotFound {
    // 未找到
}
```

#### 新代码
```go
user, err := userDO.FindByPK(ctx, userID)
if err == query.ErrRecordNotFound {
    // 未找到
}
```

### 4. 更新记录

#### 旧代码
```go
user.FullName = "Updated Name"
err := userModel.Update(ctx, user)
```

#### 新代码
```go
user.FullName = "Updated Name"
err := userDO.UpdateByPK(ctx, user)
```

### 5. 删除记录

#### 旧代码
```go
err := userModel.Delete(ctx, userID)
```

#### 新代码
```go
err := userDO.DeleteByPK(ctx, userID)
```

### 6. 条件查询

#### 旧代码
```go
users, err := userModel.SelectBuilder(ctx).
    Where(model.UserFields.Age.Eq(18)).
    Order("created_at DESC").
    FindAll()
```

#### 新代码
```go
users, err := query.user.WithDB(db).
    Where(query.userField.Age.Gte(18)).
    Order(query.userField.CreatedAt.Desc()).  // 类型安全！
    Find(ctx)
```

### 7. 批量插入

#### 旧代码
```go
users := []*model.User{...}
inserted, err := userModel.BatchInsertReturn(ctx, session, users)
```

#### 新代码
```go
users := []*query.User{...}
inserted, err := userDO.BatchCreate(ctx, users)
```

### 8. Upsert

#### 旧代码
```go
// 非零值更新
user, err := userModel.UpsertReturn(ctx, session, user)

// 全量更新
user, err := userModel.UpsertAll(ctx, session, user)
```

#### 新代码
```go
// 非零值更新
user, err := userDO.Upsert(ctx, user)

// 全量更新
user, err := userDO.UpsertAll(ctx, user)
```

### 9. 事务操作

#### 旧代码
```go
err := db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
    userModel := model.NewUserModel(conn).WithSession(session)
    user, err := userModel.InsertReturn(ctx, session, &model.User{...})
    if err != nil {
        return err
    }
    // ...
    return nil
})
```

#### 新代码
```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

userDO := query.user.WithDB(tx)
user, err := userDO.Create(ctx, &query.User{...})
if err != nil {
    return err
}
// ...

return tx.Commit()
```

### 10. 字段条件方法

#### 旧代码
```go
model.UserFields.Age.Eq(18)
model.UserFields.Age.Lt(65)
model.UserFields.Email.Like("%@example.com")
```

#### 新代码
```go
query.userField.Age.Eq(18)
query.userField.Age.Lt(65)
query.userField.Age.Between(18, 65)  // 新增！
query.userField.Email.Like("%@example.com")
query.userField.Email.ILike("alice%")  // 不区分大小写，新增！
query.userField.IsActive.IsTrue()  // 新增！
```

### 11. 排序

#### 旧代码（接受字符串，不安全）
```go
users, err := userModel.SelectBuilder(ctx).
    Order("created_at DESC", "age ASC").
    FindAll()
```

#### 新代码（类型安全！）
```go
users, err := query.user.WithDB(db).
    Order(
        query.userField.CreatedAt.Desc(),
        query.userField.Age.Asc(),
    ).
    Find(ctx)

// PostgreSQL NULL 处理
users, err := query.user.WithDB(db).
    Order(query.userField.Age.DescNullsLast()).  // 新增！
    Find(ctx)
```

## 核心优势对比

### 类型安全

#### 旧代码（部分类型不安全）
```go
// ❌ 可以传入字符串，可能拼写错误
userModel.SelectBuilder(ctx).Order("created_at DESC")

// ❌ 可以传入其他表的字段（如果不小心）
userModel.SelectBuilder(ctx).Order(model.OrderFields.Status.Asc())  // 编译通过！
```

#### 新代码（完全类型安全）
```go
// ✅ 只能使用该表字段的排序方法
query.user.WithDB(db).Order(query.userField.CreatedAt.Desc())

// ❌ 不能使用其他表的字段
query.user.WithDB(db).Order(query.orderField.Status.Asc())  // 编译错误！

// ❌ 不能传入字符串
query.user.WithDB(db).Order("created_at DESC")  // 编译错误！
```

## 迁移步骤

### 1. 备份旧代码
```bash
cp -r internal/model internal/model.backup
```

### 2. 生成新代码
```bash
make model
```

### 3. 更新导入路径
```go
// 旧
import "alpha-trade/internal/model"

// 新
import "alpha-trade/internal/query"
```

### 4. 批量替换

使用 IDE 的查找替换功能：

| 查找 | 替换 |
|-----|------|
| `model.NewUserModel(conn)` | `query.user.WithDB(db)` |
| `model.User` | `query.User` |
| `model.UserFields` | `query.userField` |
| `model.ErrNotFound` | `query.ErrRecordNotFound` |
| `.Insert(ctx,` | `.Create(ctx,` |
| `.InsertReturn(ctx, session,` | `.Create(ctx,` |
| `.FindOne(ctx,` | `.FindByPK(ctx,` |
| `.Update(ctx,` | `.UpdateByPK(ctx,` |
| `.Delete(ctx,` | `.DeleteByPK(ctx,` |
| `.BatchInsertReturn(ctx, session,` | `.BatchCreate(ctx,` |

### 5. 修复类型错误

处理编译错误，主要是：
- Order 方法改为使用字段的 Asc/Desc 方法
- 事务处理方式改变
- 去掉不需要的 session 参数

### 6. 测试

运行所有测试确保迁移成功：
```bash
go test ./...
```

## 常见问题

### Q: 可以保留旧代码吗？

A: 可以！新代码在 `internal/query` 目录，旧代码在 `internal/model` 目录，可以共存。逐步迁移即可。

### Q: 性能有变化吗？

A: 性能基本相同，都是使用 squirrel 构建 SQL。gpmg 的类型安全检查在编译期完成，运行时零开销。

### Q: 支持所有 PostgreSQL 特性吗？

A: 支持常用特性：
- ✅ 分区表
- ✅ 索引
- ✅ 数组类型
- ✅ JSONB（作为 string）
- ✅ UUID
- ✅ Decimal
- ✅ NULLS FIRST/LAST

### Q: 如何添加自定义方法？

A: 创建独立的文件（如 `users_custom.go`），扩展 DO 的功能：

```go
package query

import "context"

// 自定义方法：按邮箱域名查询
func (d *userDo) FindByEmailDomain(ctx context.Context, domain string) ([]*User, error) {
    return d.Where(userField.Email.Like("%" + domain)).Find(ctx)
}
```

### Q: 遇到问题怎么办？

1. 查看 [gpmg 文档](https://github.com/iluyuns/gpmg)
2. 查看 `internal/query/README.md`
3. 查看 `internal/query/example_usage.go`
4. 提交 Issue 到 GitHub

## 总结

gpmg 相比 gozero-pg-model-gen 的主要优势：

1. ✅ **完全类型安全**：所有 API 都是类型安全的，编译期捕获错误
2. ✅ **IDE 友好**：更好的代码提示和自动补全
3. ✅ **更简洁的 API**：链式调用更流畅
4. ✅ **更多功能**：Between、ILike、IsTrue、NULLS 处理等
5. ✅ **更少依赖**：只依赖 squirrel + lib/pq
6. ✅ **GORM Gen 风格**：更符合社区习惯

迁移过程虽然需要一些工作，但带来的类型安全和开发体验提升是值得的！
