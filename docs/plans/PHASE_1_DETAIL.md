# Phase 1: 核心领域建模 (Kernel) - 执行指南

**状态**: 🟡 进行中  
**目标**: 构建支持多标的、高精度、含完备风控的业务内核。

---

## 1. 任务清单 (Checklist)

### 1.1 基础环境与依赖
- [ ] **Go 环境**: `go mod init alpha-trade`
- [ ] **Python 环境**: 
    - 创建 `ai-agent/` 目录。
    - 安装依赖: `pip install -r ai-agent/requirements.txt`
- [ ] **NATS 服务**:
    - 启动独立 NATS Server 并开启 JetStream (`-js`)。
- [ ] 引入核心库:
    ```bash
    go get github.com/shopspring/decimal  # 高精度计算
    go get go.uber.org/zap                # 高性能日志
    go get github.com/nats-io/nats.go     # 推荐使用 NATS 作为轻量级 MQ
    ```
- [ ] 创建目录结构:
    - `internal/domain/model`
    - `internal/domain/port`
    - `internal/logic/risk`
    - `internal/infra/bus`               # 事件总线/MQ 实现

### 1.2 领域模型 (`internal/domain/model`)
**重点**: 支持多标的与费率配置。

- [ ] **`obj_money.go`**: 封装 `decimal.Decimal`。
- [ ] **`obj_symbol.go`**: 定义 `SymbolConfig` (含 `MinQty`, `MinNotional`, `FeeConfig`)。
- [ ] **`obj_market.go`**: 
    - 定义 `Tick`, `KLine`。
    - **`OrderBook` 结构**: 维护 `Bids`, `Asks` (价格->数量)，并提供 `SlippageEstimate(qty)` 方法。
- [ ] **`obj_account.go`**: 
    - `Account` 包含 `map[string]*Position`。
    - `Position` 包含 `Symbol`, `Legs` (FIFO队列)。
    - `PositionLeg` (EntryPrice, Qty, Time)。
- [ ] **`obj_settlement.go`**: 
    - 定义 `Settlement` (结算单): 记录每次平仓的 `RealizedPnL`, `Duration`, `TradeID`。
    - 必须区分 `EstimatedPnL` (预估) 和 `RealizedPnL` (实际)。
- [ ] **`obj_risk.go`**: 定义 `RiskState` (持久化状态)。

### 1.3 核心接口定义 (`internal/domain/port`)
**原则**: 接口隔离 (ISP)，现货与合约物理隔离。

- [ ] **`BaseGateway`** (基础能力):
    - `Connect(ctx)` / `Close()`
    - `GetServerTime()`: 时钟偏移修正。
- [ ] **`SpotGateway`** (现货特化):
    - `PlaceOrderFast(ctx, order)`: 仅通过 WebSocket (WS) 发送，追求极致延迟。
    - `PlaceOrderReliable(ctx, order)`: 仅通过 HTTPS 发送，作为风控或 WS 不稳定时的可靠通道。
    - `GetBalance(ctx, asset)`
    - **`RateLimiter`**: 
        - **算法选型**: 采用 **权重感知的滑动窗口 (Weight-Aware Sliding Window)**。
        - **核心机制**:
            - **严格平滑**: 拒绝令牌桶式的突发流量，确保请求在秒级均匀分布。
            - **限流反馈环**: 实时解析网关层捕获的交易所原始 Header (如 Binance 的 `X-MBX-USED-WEIGHT` 或 OKX 的 `x-ratelimit-remaining`)，并将其标准化后修正本地计数。
            - **优先级队列**: 权重接近阈值时，自动拦截开仓信号，仅保留风控平仓权限。
            - **双通道共享**: 统一管理 WS 和 HTTPS 的权重消耗。
- [ ] **`FutureGateway`** (合约特化):
    - `PlaceOrderFast(ctx, order)`: 优先 WS 通道。
    - `PlaceOrderReliable(ctx, order)`: 强制 HTTPS 通道。
    - `GetPosition(ctx, symbol)`
    - `SetLeverage(symbol, level)`
    - `SetMarginType(symbol, type)`
- [ ] **`MarketDataRepo`** (行情接口):
    - `SubscribeTick(symbol)`
    - `SubscribeKLine(symbol, interval)`
- [ ] **`EventRepo`** (宏观与外部事件):
    - **`RSSCollector`**: 实现多源 RSS 并发拉取逻辑（如 CoinDesk, CoinTelegraph, WhiteHouse）。
    - **`MacroAPIClient`**: 实现美联储 (FRED) 和 SEC 的 API 接入。
    - **`SocialMonitor`**: 基于 **RSSHub** 监控指定 X (Twitter) 名人账户。
    - **`BinanceNewsGateway`**: 实现币安公告 API 监听。
    - **`NewsFilter`**: 实现基于关键字和来源权重的过滤逻辑。
    - `FetchNews()`: 获取原始新闻并推送到 `MARKET.NEWS`。
    - **`AnalyzeSentiment(text)`**: 接入 **Gemini 3**，进行多维度情绪打分 (Score, Confidence, Keywords)。
    - `ListenRiskEvents()`: 监听全局禁号、手动暂停等信号。
- [ ] **基础设施**:
    - **RSSHub 部署**: 在 Docker 中启动私有化 RSSHub 服务。
    - **NATS 部署**: 启用 JetStream 持久化。
- [ ] **`AIProvider`** (AI 决策接口):
    - `ProcessDecision(ctx, input)`: 综合行情与新闻，利用 Gemini 3 输出决策建议。
    - `FormatPrompt(template, data)`: 提示词工程管理。
- [ ] **可靠性模块 (`internal/infra/monitor`)**:
    - **`HeartbeatMonitor`**: 监测 AI 服务存活状态，触发 Fail-safe 模式。
    - **`NATSJetStreamConfig`**: 配置持久化流与 ACK 机制。
    - **`FailSafeHandler`**: 实现决策层失联后的系统降级逻辑 (禁止开仓)。
- [ ] **`RiskRepo`**: 风控状态持久化读写。

### 1.4 风控逻辑实现 (`internal/logic/risk`)
依据 `docs/RISK_PROTOCOL.md` 实现。

- [ ] **`Manager`**: 统一入口，内部路由到不同子模块。
- [ ] **`SpotValidator`**: 实现现货资金检查。
- [ ] **`FutureValidator`**: 
    - 实现合约杠杆/保证金检查。
    - **`CheckDynamicLeverage`**: 仓位 > 10% 时强制 1x 杠杆。
    - **`CheckLiqDistance`**: 确保保留 60% 强平缓冲。
    - **`CheckRRRatio`**: 确保盈亏比 >= 1.5。
- [ ] **`Check(ctx, signal)`**: 根据 `signal.Type` 路由到对应 Validator。
- [ ] **资金流处理**:
    - `RequestWithdraw(amount decimal.Decimal) error`: 预审批提现，试算风险，调整基准。
    - `NotifyDeposit(amount decimal.Decimal) error`: 通知充值，调整额度与基准。
    - `SyncEquity()`: 兜底同步逻辑。

### 1.5 单元测试
- [ ] `risk_test.go`:
    - 测试多标的场景（如同时持有 BTC 和 ETH，检查总仓位限制）。
    - 测试手续费不足场景（预期利润 < 手续费，应拒绝）。

---

## 2. 关键变更说明

1.  **多标的支持**: `Account` 和 `RiskManager` 现在处理 map 结构的数据。
2.  **费率双轨制**: 模型中增加 `SymbolConfig.UseCustomFee`，支持回测手动指定费率。
3.  **协议对齐**: 代码实现必须严格遵守 `RISK_PROTOCOL.md` 的三层防护定义。
