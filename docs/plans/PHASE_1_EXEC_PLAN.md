# Phase 1 执行计划（按小时）- Day 1

本计划聚焦 Kernel：`internal/domain/{model,port}` 与 `internal/logic/risk`。遵循“效率优先、Security 非可选”，确保 Domain 纯净、Risk 可恢复且幂等。

---

## 现在应开发的模块（Now）
- `internal/domain/model`:
  - Money（基于 `shopspring/decimal` 的金额/数量封装）
  - Order（ID, Side, Type, Status, TIF, ClientOid 等）
  - Market（Tick, Candle，含事件时间）
  - RiskState（当日盈亏、连续亏损计数、全局状态）
- `internal/domain/port`:
  - SpotGateway, FutureGateway（下单、撤单、余额/持仓查询、设置杠杆）
  - MarketDataRepo（Ticks/KLines 订阅或拉取接口）
  - EventRepo（宏观事件/新闻信号读取）
  - RiskRepo（风控状态持久化读写）
- `internal/logic/risk`:
  - RiskManager（集中路由+聚合规则）
  - Rule: CircuitBreaker（连续亏损/MDD 熔断）
  - Rule: PositionLimit（仓位/名义价值限制）
  - Typed Errors + Unit Tests

> 约束：Domain 仅可依赖标准库与通用纯库（如 `decimal`），严禁引入 SDK/Infra/Gateway 依赖；所有 IO 通过 `port` 接口抽象。

---

## Day 1（H0-H8）时间轴

### H0 - H0.5：环境对齐与依赖就绪（不编写业务逻辑）
- 目标：
  - 锁定基础依赖：`shopspring/decimal`, `zap`, `viper`（仅添加，稍后按需引用）。
- 命令（记录以备执行）：
```bash
go get github.com/shopspring/decimal
go get go.uber.org/zap
go get github.com/spf13/viper
go mod tidy
```
- 验收：
  - `go.mod`/`go.sum` 正常更新，`go build ./...` 通过。

### H0.5 - H2：Domain Models — Money/Order/Market/RiskState
- 目标：
  - 在 `internal/domain/model` 定义以下文件（仅示例命名，按需微调）：
    - `money.go`: `type Money struct { v decimal.Decimal }` + 基础运算（Add/Sub/Mul/Div）、比较（LT/LE/GT/GE/EQ）、格式化。
    - `order.go`: 订单字段（含 `ClientOid`、`TimeInForce`、`MarketType`），状态机枚举。
    - `market.go`: `Tick`, `Candle`（事件时间与生成时间分离，回测用 EventTime）。
    - `risk_state.go`: 连续亏损计数、当日累计盈亏、MDD 观察数据、快照时间。
- 关键要求：
  - `float64` 仅可用于策略信号计算层，不进入 PnL/会计模型。
  - 时间使用 `time.Time`，回测严格走事件时间。
- 验收：
  - `internal/domain/model` 编译通过；`Money` 覆盖常见运算与零值语义。

### H2 - H3：Domain Ports — 纯接口（不可含实现）
- 目标：
  - 在 `internal/domain/port` 定义接口：
    - `SpotGateway`, `FutureGateway`: `PlaceOrder(ctx, req)`, `CancelOrder`, `GetBalance`/`GetPosition`, `SetLeverage`（合约）。
    - `MarketDataRepo`: `SubscribeTicks/SubscribeKLines` 或 `Next()` 迭代式 API（为回测留钩子）。
    - `EventRepo`: 获取宏观冷却窗口、新闻/事件流。
    - `RiskRepo`: `LoadState()`, `SaveState(delta)`，要求幂等、原子更新语义。
  - 所有方法均接收 `context.Context`，明确超时/取消语义。
- 验收：
  - 接口命名与入参出参语义自解释；无外部依赖导入。

### H3 - H5：RiskManager 骨架
- 目标：
  - 在 `internal/logic/risk`：
    - `errors.go`: 定义 `ErrRiskLimitExceeded`, `ErrCircuitBreakerOpen`, `ErrInvalidOrder` 等 Typed Errors。
    - `decision.go`: 定义 `type Decision int { Allow, Block, Reduce }` 与携带原因的 `DecisionDetail`。
    - `manager.go`: 
      - `type Manager struct { mu sync.RWMutex; repo port.RiskRepo; cfg RiskConfigSnapshot }`
      - `CheckPreTrade(ctx, req) (DecisionDetail, error)`：按规则顺序短路评估。
      - 读写状态必须经过 `RiskRepo`，本地仅做只读缓存与节流。
  - 并发安全：`sync.RWMutex` 保护内存快照；避免逃逸/减少 GC 压力。
- 验收：
  - `go build` 通过；空实现返回 `Allow` 且带 `Reason="bootstrap"`。

### H5 - H6：Rule — CircuitBreaker
- 目标：
  - 根据 `docs/RISK_PROTOCOL.md`：
    - 连续亏损 N 次、当日 MDD >= X% 触发 `Block` + `ErrCircuitBreakerOpen`。
  - 状态来源：
    - 读：`RiskRepo.LoadState(symbol/global)`；写：在实际结算/亏损事件处由调用方更新（此处仅读取并判定）。
  - 要求：
    - 判定逻辑不可产生副作用；纯函数易测试。
- 验收：
  - 单元测试涵盖：未达阈值/等于阈值/超过阈值 三类。

### H6 - H7：Rule — PositionLimit
- 目标：
  - 账户级与单标的级名义价值限制（参考 `RISK_PROTOCOL.md` 的 30%/Cash Reserve 等约束）。
  - 合约杠杆强制上限（2x）与大额单强制 1x 降档（动态限制）。
  - Fat-Finger（价格偏离/名义过大）在 `CheckPreTrade` 统一归口报错。
- 验收：
  - 覆盖：正常值、边界值、超限值；分别返回 `Allow/Reduce/Block`。

### H7 - H8：Unit Tests 与最小可运行
- 目标：
  - `internal/logic/risk/*_test.go`：
    - `Money` 基础运算/比较测试（包含负数、整除/非整除）。
    - `CircuitBreaker` 参数化测试（表驱动）。
    - `PositionLimit` 参数化测试。
  - 引入 `go.uber.org/zap` 的 `zaptest` 作为测试日志（可选）。
- 命令（记录以备执行）：
```bash
go test ./... -run Test -v
```
- 验收：
  - 单测通过；`go vet`/`staticcheck`（若已配置）无新警告。

---

## 输出物清单（Day 1 结束前）
- 目录与文件（建议命名）：
  - `internal/domain/model/{money.go, order.go, market.go, risk_state.go}`
  - `internal/domain/port/{spot_gateway.go, future_gateway.go, market_data_repo.go, event_repo.go, risk_repo.go}`
  - `internal/logic/risk/{errors.go, decision.go, manager.go, rule_circuit_breaker.go, rule_position_limit.go, *_test.go}`
- 决策流（高层）：
  - Strategy -> RiskManager(CheckPreTrade: CircuitBreaker -> PositionLimit -> …) -> OMS
- 测试：
  - CircuitBreaker/PositionLimit 均具备表驱动用例与边界覆盖。

---

## Day 2（预告，若今日进度提前）
- H8 - H10：完善 `RiskRepo` 的内存实现（Infra Mock 之前的过渡），保证状态快照与原子更新模拟。
- H10 - H12：为回测提前定义 `MarketDataRepo` 的“迭代式喂给器”接口形态与最小实现。
- H12 - H14：补齐 `pkg/telemetry` 的 `NoopTracer/Metrics` 外观，保留注入点（不侵入逻辑）。
- H14 - H16：风控规则补充（Liquidity/Slippage Guard 的契约定义），暂不接入真实 OrderBook。

---

## 风险与控制
- **Race Condition**：所有共享状态通过 `sync.RWMutex`；写少读多优先读锁。
- **GC 压力**：热路径对象避免频繁分配；优先栈上使用，必要时引入 `sync.Pool`（后续优化）。
- **Idempotency**：所有订单相关检查必须携带 `ClientOid`；`RiskRepo` 的更新采用“幂等写”语义。
- **Backtest Fidelity**：严格使用事件时间；禁用 `time.Now()`。

---

## 快速决策：此刻就开始的第一步
1) 建立 `internal/domain/model` 与 `internal/domain/port` 的文件骨架（空实现亦可，先通过编译）。
2) 在 `internal/logic/risk` 放置 `errors.go/decision.go/manager.go` 空骨架，`CheckPreTrade` 先返回 `Allow`。
3) 进入 H5 阶段，实现 `CircuitBreaker` 判定，再接 `PositionLimit` 与对应单测。

## Phase 1 执行计划（小时级排期）

**范围**: 仅涵盖 Kernel & Domain（纯净核心、不依赖外部组件），严格遵循 Domain/Logic/Infra 分层，所有会计与 PnL 使用 `shopspring/decimal`。

**你现在应该开发的地方（Now → Next）**
- **Now（优先级最高）**: `internal/domain/model` 与 `internal/domain/port` 基础模型与接口；`internal/logic/risk` 的 `RiskManager` 骨架与两条硬规则（CircuitBreaker、PositionLimit）。
- **Next**: 单元测试（Risk 核心）、`pkg/telemetry` 骨架（OTEL/Metrics 占位符），不引入外部 IO。

---

### Day 1（H0-H8）

- H0-H1（1h）: 依赖准备（仅登记/不执行 IO）
  - 目标依赖：`github.com/shopspring/decimal`、`go.uber.org/zap`、`github.com/spf13/viper`。
  - 验收：计划确认；不触达网络，后续一次性 `go get`。

- H1-H3（2h）: `internal/domain/model`
  - `Money`：`decimal.Decimal` 封装（加减乘除、格式化、Zero/Sign 辅助）。
  - `Order`：`Side/Type/Status` 枚举与基础字段（`ClientOrderID`、`Symbol`、`MarketType`、`Price`、`Quantity`…）。
  - `Market`：`Tick`、`Candle`（时间基于事件时间）。
  - `RiskState`：当日净值、连续亏损次数字段。
  - 验收：零外部依赖、编译通过、仅使用标准库+decimal。

- H3-H4（1h）: `internal/domain/port`
  - 定义 `SpotGateway`、`FutureGateway`、`MarketDataRepo`、`EventRepo`、`RiskRepo`（仅接口）。
  - 验收：接口方法签名具备 `context.Context`、`ClientOid` 幂等参数约束。

- H4-H6（2h）: `internal/logic/risk` 骨架
  - `RiskManager`：注入 `RiskRepo`，提供 `CheckPreTrade(ctx, orderCtx)` 入口。
  - Typed Errors：`ErrRiskLimitExceeded`、`ErrCircuitBreakerOpen` 等。
  - 验收：无外部 IO；仅依赖 `port` 与 `model`。

- H6-H8（2h）: 两条硬规则
  - CircuitBreaker：当日 MDD、MaxConsecutiveLosses 拦截（基于 `RiskState`）。
  - PositionLimit：`MaxSinglePositionPercent`、`MinCashReservePercent` 校验。
  - 验收：规则以纯函数实现，状态经 `RiskRepo` 读写，含边界测试样例草稿。

---

### Day 2（H8-H16）

- H8-H10（2h）: 单元测试（Risk）
  - 覆盖：正常通过、阈值刚好触发、异常输入（零或负数、极端小数）。
  - 验收：`go test` 无竞态；并发子测试验证 `RiskManager` 的读写锁保护。

- H10-H12（2h）: `pkg/telemetry` 骨架
  - 定义指标接口（PnL、Exposure、Risk Events Count）；实现空实现（no-op）+ 便于替换的构造器。
  - 验收：业务代码可注入 `Telemetry`；生产实现延后。

- H12-H14（2h）: API 约束与 Adapter 契约
  - 在 `port` 中补充 `ClientOid`、`ProtectPrice`、`timeInForce` 等参数与注释。
  - 验收：与 `DEVELOPMENT_MANUAL.md` 的 IOC/ProtectPrice 规范一致。

- H14-H16（2h）: 文档与清单
  - 更新 `docs/DEVELOPMENT_MANUAL.md` 的 Domain/Risk 片段引用与路径。
  - 更新 `docs/RISK_PROTOCOL.md` 中规则映射表（规则名 → 函数/文件位置）。
  - 验收：文档可直接指导新人 30 分钟内上手。

---

### 验收标准（Definition of Done）
- 仅 `domain` 与 `risk` 层发生变化；无 `gateway/infra` 外部依赖。
- 会计精度全链路使用 `decimal`；无 `float64` 混入（策略层另行讨论）。
- `RiskManager` 具备可恢复状态（通过 `RiskRepo`），并具备 Idempotency 输入（`ClientOid`）。
- 单测覆盖核心分支；并发路径无 Race Condition（`sync.RWMutex` 或原子计数）。

### 风险与缓解
- 精度风险：统一 `decimal`；价格/数量转换集中在 `gateway` 层做格式化。
- 并发风险：`RiskState` 内存镜像 + `RiskRepo` 持久化，读写分离与锁粒度控制。
- 范围蔓延：保持执行范围在 `domain/logic(risk)`；暂不接触 `gateway/infra`。


