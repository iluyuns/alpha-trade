# 下一步开发任务

**更新时间**: 2026-01-22  
**当前阶段**: Phase 3 实盘接入 (50%)

---

## 🎯 本周优先任务

### 1. ✅ 修复 PostgreSQL RiskRepo (完成)

**完成时间**: 2026-01-22  
**实现**:
- 统一领域模型字段映射（Symbol, InitialEquity, LastResetDate）
- JSONB + 关键字段混合存储
- Memory/Postgres 双实现通过测试

---

### 2. ✅ 实现 PostgreSQL OrderRepo (完成)

**完成时间**: 2026-01-22  
**实现**:
- 接口定义：`port.OrderRepo` (SaveOrder/GetOrder/UpdateOrderStatus/UpdateFilled/List...)
- PostgresRepo：基于 orders 表，支持幂等保存
- MemoryRepo：回测专用，线程安全
- 所有测试通过

---

### 3. 实现 WebSocket 行情订阅 🔴

**目标**: 实时接收 Binance 行情  
**功能**: Kline/Ticker 订阅、自动重连、心跳检测  
**预计**: 6h

---

### 4. 集成测试框架 🔴

**目标**: 端到端验证  
**场景**: Binance API、数据库持久化、完整交易流程  
**预计**: 4h


## 📋 后续任务

### 5. Redis 缓存层 (下周)
- 价格/订单簿/余额缓存
- **预计**: 4h

### 6. 订单状态同步 (下周)
- WebSocket User Data Stream
- 定期轮询补偿
- **预计**: 6h

### 7. 监控告警 (下周)
- Prometheus 指标
- 关键业务监控
- **预计**: 4h

---

## 📝 技术债务

**高优先级**:
1. ✅ 修复 RiskState 字段映射
2. ✅ OrderRepo 持久化实现
3. ⚪ SpotGateway.GetOrder() 改进
4. ⚪ 错误处理统一
5. ⚪ 连接池配置

**中优先级**:
5. ⚪ GPMG 生成 query 代码
6. ⚪ 批量查询优化
7. ⚪ 事务管理

---

## 🎯 成功标准

- [x] PostgreSQL 持久化稳定 (RiskRepo + OrderRepo)
- [ ] WebSocket 实时行情正常
- [ ] 订单状态同步无遗漏
- [ ] 集成测试通过率 > 95%
- [ ] 代码覆盖率 > 60%

**预计完成**: 2026-01-26
