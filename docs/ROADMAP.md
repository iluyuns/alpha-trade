# Alpha-Trade 项目开发路线图 (Roadmap)

本文档用于跟踪项目整体开发进度，确保按照“核心优先、回测驱动、实盘落地”的节奏有序进行。

---

## ⏱ 下一步与执行入口

- 立即执行入口: 详见 [NEXT_STEPS.md](../NEXT_STEPS.md)（Phase 3 实盘接入）。
- 当前应开发区域: `internal/infra/`、`internal/gateway/binance/`、`internal/logic/oms`、`internal/infra/risk`。

---

## 📅 总体进度概览

| 阶段 | 核心目标 | 状态 | 预计周期 | 关键产出 |
| :--- | :--- | :--- | :--- | :--- |
| **Phase 1** | **核心领域建模 (Kernel)** | ✅ 已完成 | 1 天 | 领域模型(Decimal), 风控核心(带状态), 基础架构 |
| **Phase 2** | **回测系统构建 (Simulator)** | ✅ 已完成 | 1 天 | Mock交易所, CSV加载器, 波动策略实现, 回测报告 |
| **Phase 3** | **实盘接入适配 (Real World)** | 🔄 进行中 (90%) | 3-5 天 | Binance API适配, 数据库持久化, 信号执行 |
| **Phase 4** | **生产环境部署 (Production)** | ⚪ 待开始 | 2-3 天 | Docker镜像, 监控面板(Grafana), 告警机器人 |

---

## 🛠️ Phase 1: 核心领域建模 (Kernel)

**目标**: 建立不依赖任何外部组件的纯净核心。无论回测还是实盘，都完全复用此层代码。

### 1.1 项目初始化 ✅
- [x] 初始化 `go.mod` (Module 已存在)
- [x] 创建标准目录结构 (`cmd`, `internal`, `docs` 已存在；配置目录采用 `config/`)
- [x] 引入基础依赖 (`shopspring/decimal`, `zap`, `viper`)
- [ ] **安全基建**: 引入 `go-webauthn` 并设计用户表结构 (支持 Passkeys)。（数据库表已存在；SDK 集成与校验链路待落地）

### 1.2 领域模型定义 (`internal/domain/model`) ✅
- [x] **Money**: 封装 `decimal.Decimal`，处理金额/数量计算。
- [x] **Order**: 定义订单结构 (ID, Side, Type, Status)。
- [x] **Market**: 定义行情结构 (Candle, Tick)。
- [x] **RiskState**: 定义风控状态 (连续亏损次数, 当日盈亏)。

### 1.3 核心接口定义 (`internal/domain/port`) ✅
- [x] **`SpotGateway`**: 现货交互接口 (发单, 查余额)。
- [x] **`FutureGateway`**: 合约交互接口 (发单, 查持仓, 设杠杆)。
- [x] `MarketDataRepo`: 行情流接口 (Tick/KLine)。
- [x] **`EventRepo`**: 宏观事件/新闻/风控信号接口。
- [x] **`RiskRepo`**: 风控状态持久化读写接口。

### 1.4 风控核心逻辑 (`internal/logic/risk`) ✅
- [x] **Manager**: 风控管理器，集成 `RiskRepo`。
- [x] **Rule: CircuitBreaker**: 熔断机制 (依赖持久化状态)。
- [x] **Rule: PositionLimit**: 仓位限制。
- [x] **Tests**: 编写核心风控的单元测试 (Unit Tests)。

### 1.5 可观测性基建 (Observability)
- [ ] **Tracing**: 搭建 `pkg/telemetry` 基础骨架。
- [x] **Metrics**: 定义核心业务指标 (PnL, Exposure) - Phase 3 已完成。

---

## 🔬 Phase 2: 回测系统构建 (Simulator)

**目标**: 在零成本环境下验证策略逻辑与风控规则。

### 2.1 模拟基础设施 (`internal/gateway/backtest`)
- [x] **MockExchange**: 内存模拟交易所 (实现 `ExchangeRepo`)。
- [x] **MockEventSource**: 模拟新闻/事件发生。
- [x] **DataLoader**: CSV 历史数据读取器。

### 2.2 策略实现 (`internal/logic/strategy`)
- [x] **BaseEngine**: 策略基类。
- [x] **VolatilityStrategy**: 实现波动监控策略。

### 2.3 回测运行器 (`cmd/backtest`)
- [x] 编写 `main.go` 组装 Mock 组件。
- [x] 输出回测统计报告 (收益率, 最大回撤, 夏普比率)。

---

## 🔌 Phase 3: 实盘接入适配 (Real World)

**目标**: 将模拟组件替换为真实交易所适配器，并添加持久化。

### 3.1 外部适配器 (`internal/gateway`)
- [x] **BinanceAdapter**: 基于 `go-binance` 封装 REST/WS (SpotClient + WSClient)。
- [ ] **NewsAdapter**: 接入新闻源 API (如 CryptoCompare/Calendar) - Phase 4+。

### 3.2 基础设施 (`internal/infra`)
- [x] **Persistence**: 实现 PostgreSQL 存储交易记录 (OrderRepo + RiskRepo)。
- [x] **Redis**: 实现 `RiskRepo` 的 Redis 版本 (状态持久化，支持 TTL)。

### 3.3 订单管理 (`internal/logic/oms`)
- [x] **OrderManager**: 订单状态同步与生命周期管理 (集成风控检查)。

### 3.4 集成测试
- [x] **端到端测试框架**: 验证 Strategy -> RiskManager -> OMS -> Gateway -> OrderRepo 完整链路。

### 3.5 可观测性
- [x] **Prometheus Metrics**: 基础指标集成 (订单、风控、PnL、延迟)。

---

## 🚀 Phase 4: 生产环境部署 (Production)

**目标**: 提高系统的健壮性与可观测性。

### 4.1 监控与告警
- [x] 集成 Prometheus (`/metrics`) - 基础指标已实现。
- [ ] 配置 Grafana Dashboard (实时风控状态可视化)。

### 4.2 容器化
- [x] `Dockerfile` (多阶段构建，Alpine 基础镜像) ✅
- [x] `docker-compose.yml` (应用服务集成) ✅

### 4.3 告警通知
- [ ] 集成 Telegram/Lark Bot。

