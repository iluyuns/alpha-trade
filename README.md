# Alpha-Trade

量化交易系统 - 核心优先、回测驱动、实盘落地

**当前版本**: v0.3.0-alpha  
**最后更新**: 2026-01-22

## 📊 项目状态

```
Phase 1: 核心领域建模    ████████████████████  100% ✅
Phase 2: 回测系统构建    ████████████████████  100% ✅
Phase 3: 实盘接入适配    ██████░░░░░░░░░░░░░░   30% 🔄
Phase 4: 生产环境部署    ░░░░░░░░░░░░░░░░░░░░    0% ⚪
```

### Phase 3 当前进展
- ✅ Binance REST API 客户端 (100%)
- 🔄 PostgreSQL 风控持久化 (50%)
- ⚪ PostgreSQL OrderRepo
- ⚪ WebSocket 行情订阅
- ⚪ Redis 缓存层
- ⚪ 订单状态同步

---

## 快速开始

### 回测运行

```bash
# 编译
go build -o bin/backtest ./cmd/backtest

# 运行回测
./bin/backtest \
  -csv testdata/sample_btc.csv \
  -symbol BTCUSDT \
  -threshold 0.02 \
  -capital 10000
```

### 测试

```bash
# 运行所有测试
go test ./internal/... -v

# 运行特定模块
go test ./internal/strategy/... -v
```

---

## 核心模块

### Domain (`internal/domain/`)
- `model/` - 领域模型（Money/Order/Market/RiskState）
- `port/` - 接口定义（Gateway/Repo）

### Risk (`internal/logic/risk/`)
- `manager.go` - 风控管理器
- `rule_circuit_breaker.go` - 熔断器
- `rule_position_limit.go` - 仓位限制

### Strategy (`internal/strategy/`)
- `engine.go` - 策略引擎
- `simple_volatility.go` - 波动策略

### Gateway (`internal/gateway/`)
- `mock/` - 模拟交易所（回测用）
- `binance/` - Binance REST API 客户端 ⭐ **NEW**

### Infrastructure
- `internal/infra/risk/` - 风控仓储（内存/PostgreSQL）
- `internal/backtest/loader/` - CSV数据加载器

---

## 技术栈

- **语言**: Go 1.25+
- **精度**: shopspring/decimal
- **交易所**: binance-connector-go v0.8.0 ⭐
- **数据库**: PostgreSQL 14+ (实盘)
- **缓存**: Redis 7+ (计划中)
- **日志**: go.uber.org/zap
- **配置**: spf13/viper

---

## 📚 文档索引

### 核心文档
- [下一步任务](NEXT_STEPS.md) - 本周开发计划
- [开发路线图](docs/ROADMAP.md) - 整体规划
- [开发手册](docs/DEVELOPMENT_MANUAL.md) - 架构设计与规范
- [风控协议](docs/RISK_PROTOCOL.md) - 爷叔风控规则
- [安全协议](docs/SECURITY_PROTOCOL.md) - 系统安全基线
- [文档规范](docs/DOC_RULES.md) - 文档编写与维护规范 ⭐

### 历史记录
- [Phase 2 总结](docs/archive/PHASE_2_SUMMARY.md)
- [Phase 3 总结](docs/archive/PHASE_3_SUMMARY.md)
- [修复总结](docs/archive/FIXES_SUMMARY.md)

---

## 项目结构

```
alpha-trade/
├── cmd/
│   └── backtest/          # 回测运行器
├── internal/
│   ├── domain/            # 领域层
│   │   ├── model/         # 领域模型
│   │   └── port/          # 接口定义
│   ├── logic/
│   │   └── risk/          # 风控逻辑
│   ├── strategy/          # 策略引擎
│   ├── gateway/
│   │   └── mock/          # 模拟交易所
│   ├── infra/
│   │   └── risk/          # 风控基础设施
│   └── backtest/
│       └── loader/        # 数据加载器
├── testdata/              # 测试数据
└── docs/                  # 文档
```

---

## License

Private Project - All Rights Reserved
