# Alpha-Trade 风控协议 (Risk Management Protocol)

**状态**: 生效中 (Active)
**级别**: 核心安全规范 (Critical)

本文档重新定义了系统的风控体系。基于量化交易的最佳实践，我们将风控分为 **L1 (系统级)**、**L2 (账户级)** 和 **L3 (策略级)** 三层防护网。

---

## 1. 风控哲学 (Risk Philosophy)

1.  **生存第一**: 在本金归零之前，任何策略都毫无意义。风控模块有权在不通知策略的情况下**拒绝订单**或**强制平仓**。
2.  **分层防御**: 即使策略逻辑完全出错，底层的硬风控也能守住资金底线。
3.  **悲观假设**: 计算可用资金时，总是假设手续费是最高的；计算盈亏时，总是假设滑点是最差的。
4.  **状态持久化**: 风控状态（如今日亏损额）必须持久化存储，**不可因重启而重置**。

---

## 2. L1: 系统级风控 (System Hard Stops)

**触发后果**: **全系统停机 (Halt)**，取消所有挂单，仅允许平仓请求，发送最高级别告警。

### 2.1 熔断机制 (Circuit Breaker)
*   **连续亏损熔断**: 全局连续亏损交易达到 **N次** (e.g., 5次)。
    *   *逻辑*: 既然连输 5 次，说明市场环境变了或模型失效，必须人工介入检查。
*   **回撤熔断**: 当日累计净亏损达到总资金的 **X%** (e.g., 5%)。
    *   *逻辑*: 锁定当日最大回撤，避免情绪化交易。

### 2.2 宏观冷却 (Macro Event Cooling)
*   **规则**: 在重大财经事件（如美联储议息、CPI发布）前后 **N小时** 内，禁止开新仓。
*   **实现**: 依赖 `EventGateway` 获取日历数据，风控模块检查当前时间戳是否在黑名单窗口内。

### 2.3 技术异常保护
*   **API 错误率**: 过去 1 分钟内 API 失败率 > 20%。
*   **数据延迟**: 行情数据滞后超过 3 秒 (WebSocket 拥堵)。

### 2.4 外部资金流变动 (External Flow Handling)
*   **资金围栏 (Capital Fence)**:
    *   引入 `AllocatedCapital` 配置。系统计算可用资金时，取 `Min(ExchangeEquity, AllocatedCapital)`。
*   **资金变动预审批 (Capital Change Pre-Approval)**:
    *   **减少资金 (Withdraw)**: 必须先调用 `Risk.RequestWithdraw(amount)`。
        *   系统试算：`NewEquity` 是否满足 `Margin + Reserve` 要求。
        *   若通过：自动扣减 `AllocatedCapital` 并下调 `DailyStartEquity`。
        *   若失败：拒绝请求，提示强平风险。
    *   **增加资金 (Deposit)**: 必须先调用 `Risk.NotifyDeposit(amount)`。
        *   系统增加 `AllocatedCapital` 并上调 `DailyStartEquity`。
*   **净值重定标 (Equity Re-baselining)**:
    *   作为**兜底机制**，当检测到未经审批的权益剧烈变化时，触发“异常资金变动”告警，并尝试自动对齐（但不保证不熔断）。

### 2.5 网络与数据完整性 (Network & Data Integrity)
*   **数据时效性 (Stale Data Check)**:
    *   策略计算前校验: `Now - MarketData.Time > 3s`?
    *   **动作**: 拒绝生成信号，Gateway 触发重连。
*   **死人开关 (Deadman Switch)**:
    *   系统需定期 (e.g. 10s) 探测内部健康状态。
    *   若发现核心进程死锁或崩溃，外部独立监控脚本 (Watchdog) 必须立即调用交易所 API **撤销所有挂单**。

### 2.6 交易所交互安全 (Exchange Interaction)
*   **本地限流 (Local Rate Limiter)**:
    *   Gateway 必须维护本地令牌桶，严格遵守交易所 Rate Limit 规则。
    *   **动作**: 在请求耗尽权重前主动阻塞，严禁触发 IP Ban。
*   **时钟同步 (Time Sync)**:
    *   启动时及每小时检测: `Abs(LocalTime - ExchangeTime) > 1s`?
    *   **动作**: 警告并自动计算 TimeOffset 修正请求头，若偏差过大 (>5s) 则停机。
*   **幂等性 (Idempotency)**:
    *   所有非查询类请求必须携带 `ClientOrderID`，防止网络超时重试导致的重复下单。

### 2.7 AI 治理与权限 (AI Governance)
*   **默认原则**: AI (LLM/Sentiment Analysis) 默认为 **"Advisor Mode" (仅建议)**，无权直接操作账户。
*   **权限分级**:
    *   **Level 0 (Notify)**: AI 识别风险 -> 发送告警 -> 人工决策 (默认)。
    *   **Level 1 (Defensive)**: 允许 AI 触发 "暂停开仓" 或 "减仓" (只减不加)。
    *   **Level 2 (Autonomous)**: 允许 AI 执行 "一键清仓" 或 "反向开仓" (需物理开关开启 + 二次确认)。
*   **置信度阈值**: AI 信号必须附带 Confidence Score (>0.9) 才能触发 L1/L2 操作。

---

## 3. 告警与通知协议 (Notification Hierarchy)

为了防止告警疲劳并确保极速响应，系统必须实现分级推送逻辑。

| 级别 | 定义 | 链路 | 响应要求 |
| :--- | :--- | :--- | :--- |
| **INFO** | 策略成交、心跳存活、AI 情绪波动 | 飞书机器人 | 异步审计 |
| **WARN** | API 延迟 > 2s、网络重连、滑点 > 0.3% | 飞书 + Bark (普通) | 1 小时内检视环境 |
| **CRITICAL** | **L1/L2 熔断、保证金不足、非预审批提现** | **飞书 + Bark (紧急) + 短信/语音** | **立即介入 (365x24h)** |

### 3.1 核心要求
*   **Bark 紧急模式**: 针对 CRITICAL 级别，必须使用 `isCritical=1` 绕过静音开关。
*   **私有化部署 (Self-Host)**: 告警模块必须支持自定义 Bark 推送节点 URL，优先使用自建 Docker 节点以确保隐私与实时性。
*   **短信脱敏**: 为绕过大陆短信审核，CRITICAL 级别短信必须使用中性化编码模板 (e.g. 状态码 911)。
*   **心跳静默检测**: 若 1 小时未收到 INFO 级别心跳消息，则自动判定系统崩溃，触发备用链路告警。

---

## 4. L2: 账户级风控 (Account Hygiene)

**触发后果**: **拒绝开仓 (Reject)**，但不影响现有持仓。

### 3.1 杠杆与仓位管理
*   **Hard Leverage Limit (硬杠杆限制)**: 
    *   合约交易杠杆 **绝对上限 2x**。
*   **Dynamic Leverage Limit (动态杠杆/阶梯风控)**:
    *   **大仓位保护**: 若单笔订单价值 > 账户总权益的 **10%**，则强制该笔交易只能使用 **1.0x 杠杆**。
    *   *逻辑*: 既然单笔占比较大，必须放弃杠杆以确保 100% 的价格波动容错，防止“重仓+杠杆”导致的快速毁灭。
*   **Cash Reserve**: 必须保留 **5%** 的现金作为安全垫，用于支付手续费和应对极端行情滑点。

### 3.2 资金分散度
*   **单一标的限制**: 单个 Symbol 的持仓价值不得超过总权益的 **30%**。
    *   *目的*: 防止单币种黑天鹅（如 LUNA 归零）摧毁整个账户。

### 3.3 保证金模式强制 (Margin Mode Enforcement)
*   **Force Isolated (强制逐仓)**:
    *   系统初始化或下单前，必须校验当前交易对是否为 **ISOLATED (逐仓)** 模式。
    *   若检测到 CROSS (全仓) 模式，系统必须**拒绝交易**或**自动切换为逐仓**。
    *   *逻辑*: 物理隔离风险，防止单点代码故障导致全账户资金归零 (Firewall)。

---

## 4. L3: 交易/订单级风控 (Pre-Trade Validation)

**触发后果**: **拒绝订单 (Reject)**，记录 Warning 日志。

### 4.1 强制止损与盈亏比 (Mandatory Stop & R/R)
*   **强制附带止损 (Must Have Stop-Loss)**:
    *   所有开仓请求**必须**包含明确的 `StopLossPrice`。
    *   若策略未指定，风控层直接拒单。
*   **盈亏比倒挂检查 (Inverted R/R Check)**:
    *   规则: `PotentialProfit / PotentialLoss >= 1.5`。
    *   *逻辑*: 严禁“冒 5% 风险去赚 1% 利润”。若 `TP_Dist / SL_Dist < 1.5`，拒单。

### 4.2 强平防卫距离 (Liquidation Buffer)
*   **规则**: 止损价必须远离强平价，保留至少 **60%** 的安全缓冲区。
*   **公式**: `Abs(StopPrice - EntryPrice) <= Abs(LiqPrice - EntryPrice) * 0.4`
    *   *(即：止损距离只能占用强平距离的 40%，剩余 60% 用于应对极端滑点)*。
*   *场景*: 强平价 100，开仓价 110 (距离10)。止损价最低只能设在 106 (消耗4)，必须保留 100-106 这段区间作为防穿仓缓冲。

### 4.3 胖手指检查 (Fat Finger Check)
*   **价格偏离**: 订单价格与当前市场最新成交价 (Last Price)偏差超过 **5%**。
*   **名义价值过大**: 单笔订单价值超过 **$50,000** (可配置)。

### 4.4 流动性与滑点控制 (Liquidity & Slippage)
*   **深度检查 (Depth Check)**:
    *   下单前必须通过本地 OrderBook 计算：`1% 深度内的累计价值 >= 订单名义价值 * 5`。
    *   *逻辑*: 确保单笔订单不占用盘口核心深度的 20% 以上，防止产生极端冲击成本。
*   **滑点预估保护 (Slippage Guard)**:
    *   规则: `(预估平均成交价 - 当前市价) / 当前市价 <= 0.5%` (根据币种流动性可调)。
    *   *动作*: 若预估滑点超过阈值，系统必须拒绝市价单，转为 **限价单 (Limit Order)** 或 **算法拆单 (TWAP)**。

### 4.2 费率保护 (Fee Protection)
*   **规则**: 计算预期盈利时，必须扣除 **双倍手续费** (开仓+平仓)。
*   **公式**: 如果 `ExpectedProfit < OrderValue * (MakerFee + TakerFee) * 1.5`，则该交易无利可图，直接过滤。

---

## 5. 风控配置模型

```go
type RiskConfig struct {
    // L1: System
    MaxDailyDrawdownPercent decimal.Decimal // e.g. 0.05 (5%)
    MaxConsecutiveLosses    int             // e.g. 5
    
    // L2: Account
    MaxGlobalLeverage       decimal.Decimal // e.g. 1.0 (Spot)
    MinCashReservePercent   decimal.Decimal // e.g. 0.05
    
    // L3: Strategy/Symbol
    MaxSinglePositionPercent decimal.Decimal // e.g. 0.30
    MaxSlippageTolerance     decimal.Decimal // e.g. 0.005 (0.5%)
}
```

---

## 6. 应急操作流程 (Emergency Ops)

1.  **Kill Switch (一键清仓)**:
    *   运维指令，忽略所有成本，以市价单 (Market Order) 强平所有持仓。
    *   场景: 交易所被黑、私钥泄露、代码逻辑死循环。
2.  **Degraded Mode (降级模式)**:
    *   当行情延迟变大时，自动禁止 Limit Order，仅允许 Market Order 平仓。

---

**修订历史**:
- v1.0: 初始版本，基于通用量化经验重构。
- v1.1: 增加宏观冷却与费率保护细节。

