# Alpha-Trade 爷叔风控协议 (Ye Shu Risk Protocol)

**状态**: 绝对执行 (Strictly Enforced)
**级别**: 系统最高真理 (The Supreme Truth)
**执行官**: `RiskManager` (即“爷叔”)

> “先要学会输，才能学会赢。” —— 爷叔

本文档定义了 Alpha-Trade 系统的核心生存逻辑。在黄河路的数字丛林里，**爷叔（RiskManager）** 是那个在所有人疯狂时强行拉下电闸的人。他不是为了帮你赚钱，他是为了保住你的派头和底牌。

---

## 1. 爷叔的生存哲学 (Risk Philosophy)

1.  **底牌第一 (Bottom Line)**：在你的本金归零之前，任何所谓的 Alpha 都是幻象。爷叔有权在不通知任何人的情况下，直接**收回头寸（强平）**或**封锁柜台（拒单）**。
2.  **派头要稳 (Positioning)**：赚多少钱是运气，亏多少钱是规矩。即使你的“阿宝策略”逻辑全对，只要触碰了爷叔的红线，一律按死。
3.  **悲观主义 (Pessimism)**：爷叔从不相信奇迹。算账时，手续费按最高算，滑点按最差算，行情按最极端算。
4.  **死记账 (Persistence)**：爷叔的账本（风险状态）是刻在石头上的，**重启系统也抹不掉**。想通过重启来重置亏损限额？门都没有。

---

## 2. 价格参考标准 (Price Reference Standards)

为了防止单点异常行情（插针）误触发风控或导致开仓异常，系统执行以下价格引用准则：

| 价格类型 | 定义 | 系统应用场景 |
| :--- | :--- | :--- |
| **最新成交价 (Last Price)** | 当前交易所的最后一笔成交价 | 订单执行、紧急平仓触发、市价单滑点估算。 |
| **标记价格 (Mark Price)** | 经过平滑处理的合约参考价 | **未实现盈亏计算、账户健康度评估、风控止损触发、强平防卫。** |
| **指数价格 (Index Price)** | 多个主流交易所现货价格的加权平均 | **开仓价格偏离度检查 (Fat-Finger Guard)。** |

---

## 3. L1 级：爷叔的“电闸” (The Kill Switch - System Level)

**触发后果**: **全系统强制休克 (Halt)**。撤销黄河路上所有的挂单，拒绝一切新开仓，仅允许止损平仓，发送 CRITICAL 级告警。

### 3.1 熔断机制：拉电闸 (Circuit Breaker)
*   **累计亏损熔断**：当日累计净亏损达到总资金的 **5%**（这是爷叔给你留的底线）。
    *   *规矩*：既然今天手气背，就别在黄河路混了，回家睡觉。
*   **连续输单熔断**：全局连续亏损交易达到 **5次**。
    *   *规矩*：连输五次，说明你的“阿宝策略”已经和市场对不上了，必须停下来检修。

### 3.2 宏观冷却：601 战役保护 (Macro Event Cooling)
*   **规矩**：在重大数据发布（如 CPI、美联储议息）前后 1 小时内，爷叔锁死柜台，不准开仓。
*   **实现**: 依赖 `EventGateway` 获取日历数据，风控模块检查当前时间戳是否在黑名单窗口内。

### 3.3 技术异常与死人开关
*   **API 异常**: 过去 1 分钟内 API 失败率 > 20% 或行情延迟 > 3s。
*   **死人开关 (Deadman Switch)**: 发现核心进程死锁，外部 Watchdog 必须立即撤销所有挂单。

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
    *   Gateway 必须维护本地**滑动窗口 (Sliding Window)**，严格遵守交易所 Rate Limit 规则。
    *   *说明*: 弃用令牌桶，防止突发流量 (Burst) 触发交易所的 IP 权重封禁。
    *   **动作**: 在请求耗尽权重前主动阻塞。
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

## 4. L2 级：爷叔的“账本” (Portfolio Hygiene)

**触发后果**: **拒单 (Reject)**。不影响你手里已有的头寸，但想加码？没门。

### 4.1 派头限制：杠杆与仓位 (Leverage & Sizing)
*   **铁律：2x 杠杆上限**。在爷叔这里，10 倍、20 倍那是赌博，不是做生意。
*   **单一标的限制**：单个 Symbol 的持仓价值不得超过总权益的 **30%**。
    *   *规矩*：别把鸡蛋都放一个篮子里，LUNA 归零的教训，爷叔记得比你清。
*   **备用金要求 (Cash Reserve)**：必须保留 **5%** 的现金，用于支付手续费和应对极端行情滑点。

### 4.2 逐仓模式强制 (Force Isolated)
*   **规矩**：开仓前必须检查是否为 **逐仓 (ISOLATED)** 模式。全仓 (CROSS) 模式在爷叔这儿是不准进场的，物理隔离风险是第一要素。

---

## 5. L3 级：爷叔的“门禁” (Pre-Trade Gatekeeper)

**触发后果**: **拦单 (Block)**。记录警告，拒绝该笔订单进入网关。

### 5.1 止损与盈亏比：没后路不准进场
*   **强制止损 (Must Have SL)**：所有开仓请求**必须**附带 `StopLossPrice`。
*   **盈亏比要求**：预期利润必须至少是风险的 **1.5 倍**（即 `TP_Dist / SL_Dist >= 1.5`）。严禁“冒 5% 风险去赚 1% 利润”。

### 5.2 强平防卫距离 (Liquidation Buffer)
*   **规矩**：止损价必须远离强平价。止损位最多只能占用 40% 的强平空间，剩下的 60% 是留给爷叔应对黑天鹅的“安全垫”。

### 5.3 胖手指检查 (Fat Finger Check)
*   **价格偏离**：订单价偏离市价超过 **5%**，拒单。
*   **名义价值**：单笔订单价值超过 **$50,000** (默认)，拒单。

---

## 6. 爷叔的“规矩”代码实现 (The Interceptor Contract)

为了确保“爷叔”能管住“阿宝”，所有核心代码必须遵循以下 **拦截器契约 (The Interceptor Contract)**：

1.  **强制路由**：在任何 `Strategy` 生成信号后，`OMS` 在调用 `Gateway` 之前，**必须同步调用** `RiskManager.CheckOrder()`。
2.  **原子决策**：风控拦截操作必须是原子的，且在内存中完成以确保极速（< 1ms）。
3.  **解释权归属**：当策略逻辑与风控规则冲突时，**爷叔拥有最高裁定权**，网关只接受通过风控校验的订单。

### 4.4 流动性与滑点控制 (Liquidity & Slippage)
*   **深度检查 (Depth Check)**:
    *   下单前必须通过本地 OrderBook 计算：`1% 深度内的累计价值 >= 订单名义价值 * 5`。
    *   *逻辑*: 确保单笔订单不占用盘口核心深度的 20% 以上，防止产生极端冲击成本。
*   **滑点预估保护 (Slippage Guard)**:
    *   **规则**: 严禁无保护的市价单。所有市价单必须在 API 层封装为带有价格保护的 **IOC (Immediate-or-Cancel)** 限价单。
    *   **价格上限 (Long/Buy)**: `ExecutionPrice <= SignalPrice * (1 + SlippageTolerance)`。
    *   **价格下限 (Short/Sell)**: `ExecutionPrice >= SignalPrice * (1 - SlippageTolerance)`。
    *   **动作**: 若瞬时滑点超过阈值，交易所会部分成交或全部撤单，系统不再二次追单。

### 4.6 追高与踏空保护 (Anti-Chasing Logic)

针对“ V 型反转”或“插针后暴拉”场景，系统执行以下硬性拦截：

*   **信号价格有效区间 (Price Valid Zone)**:
    *   系统记录信号触发时的 `SignalPrice`。
    *   开仓请求到达网关时，若 `abs(CurrentPrice - SignalPrice) / SignalPrice > MaxDeviation (default 0.5%)`，即使未达到策略止损，也必须判定为“信号已失效”，严禁追单。
*   **信号时效检查 (Latency Kill-switch)**:
    *   若 `Now() - SignalTimestamp > 500ms`，视为执行延迟过高，自动作废该次开仓指令。

### 4.2 费率与持有成本保护 (Fee & Carrying Cost Protection)
*   **规则**: 计算预期盈利时，必须扣除 **双倍手续费** (开仓+平仓)；对于 **合约 (Future)**，还必须额外扣除 **预估资金费率成本**。
*   **多资产折算 (Fee Normalization)**:
    *   若手续费并非以 Quote Asset (如 USDT) 结算（例如使用 BNB 抵扣），系统**必须**读取 `BNBUSDT` 的实时价格，将手续费折算为 USDT 后再参与 PnL 减法计算，严禁直接数值相减。
*   **公式**: 
    *   **Spot**: `ExpectedProfit < (OrderValue * (MakerFee + TakerFee)) * 1.5`
    *   **Future**: `ExpectedProfit < (OrderValue * (MakerFee + TakerFee) + EstimatedFundingCost) * 1.5`
*   **异常资金费率拦截 (Future Only)**: 若当前标的资金费率 (Funding Rate) 绝对值超过 **0.1%** (单次结算)，风控模块需强制检查持仓时长预估，防止因费率损耗导致本金快速缩水。

### 4.5 开仓增强与确认协议 (Entry Enhancement & Confirmation)

为了提高开仓质量，系统对 L3 级开仓指令应用以下增强逻辑：

*   **开仓公平价校验 (Fair Value Check)**:
    *   **规则**: `abs(LastPrice - MarkPrice) / MarkPrice <= 0.3%`。
    *   **动作**: 若偏离度过高，系统强制将市价单 (Market) 降级为限价单 (Limit)，价格挂在 `MarkPrice * (1 + 0.1%)`，防止追涨杀跌。
*   **信号确认窗口 (Signal Confirmation)**:
    *   **逻辑**: 突破类策略触发后，引入 **动态观察窗 (Adaptive Window)**。
    *   **配置**: 由策略根据其频率属性定义（默认：中频策略 3s，高频策略 100ms）。
    *   **动态调整**: 若当前 ATR > 均值 2 倍，系统自动将观察窗延长 50%，以应对极端波动下的假突破。
    *   **执行条件**: 观察期内 Mark Price 必须持续维持在触发位以上，若跌回则视为假突破，作废信号。
*   **波动率头寸缩放 (Volatility Sizing)**:
    *   **计算**: `PositionSize = AccountRiskAmount / (ATR * N)`。
    *   **目的**: 波动大时买少点，止损远点；波动小时买多点，止损近点。确保每笔交易的“期望风险额”恒定。

---

## 5. 亏损平仓专项处理 (Loss-Exit Execution)

亏损平仓被定义为 **“紧急防御动作”**，其执行逻辑不同于常规开仓。

### 5.1 执行优先级
*   **通道隔离**: 亏损平仓指令优先占用网关的高速通道。
*   **指令类型**: 
    *   默认使用 **Market Order**。
    *   若使用 Limit Order，必须附带 **Price Buffer** (例如：卖出价 = Bid_Price * 0.99)，确保即时成交。

### 5.2 流动性保护 (Liquidity Safety)
*   **冲击成本预警**: 若平仓单价值 > 盘口 1% 深度的 20%，系统必须强制切换为 **TWAP 拆单模式**，在 N 秒内完成平仓，禁止单笔大额市价单直接撞单。

### 5.3 熔断联动 (Circuit Breaker Linkage)
*   **连续亏损熔断 (L1-CB)**: 亏损平仓后，`MaxConsecutiveLosses` 计数器加 1。达到阈值时触发全局停机。
*   **权益强制校准**: 每一笔平仓结算后，必须立即通过 `Account.SyncEquity()` 刷新可用保证金，防止资产净值下降导致的后续违规。

---

## 6. 防扫单与防插针优化 (Anti-Whipsaw Optimization)

为了减少因市场瞬间插针导致的“无辜止损”，系统引入以下平滑机制：

### 6.1 标记价格优先原则
*   **规则**: L3 级止损指令的触发判定**必须**使用 **Mark Price**。
*   **理由**: 过滤掉单一交易所因盘口空虚产生的 Last Price 虚假插针。

### 6.2 确认延迟 (Confirmation Buffer)
*   **逻辑**: 当 Mark Price 首次穿过止损位时，系统不立即下单，而是开启一个 **Time Window (e.g., 1.5s)**。
*   **执行条件**: 只有在窗口期结束时，价格仍未回调至止损位上方，才触发执行。
*   **适用性**: 仅适用于非极端行情。若价格瞬间跌幅超过 3%，则绕过确认期立即平仓。

### 6.3 波动率自适应止损 (ATR-Based SL)
*   **动态调整**: 策略生成的 `StopLossPrice` 应参考当前标的的 **ATR 指标**。
*   **风控限制**: 即使策略请求更紧的止损，风控层也会根据当前 1min 波动率强制保留一个“最小安全垫”，防止在震荡区间被频繁扫单。

---

## 7. 绩效风控与胜率管理 (Performance-Based Risk Control)

系统不仅关注单笔订单的价格，还通过实时统计策略的**胜率 (Win Rate)** 和 **期望值 (Expectancy)** 来实现动态风险调整。

### 7.1 核心评价指标
*   **实时胜率 (Real-time Win Rate)**: 基于滑动窗口 (默认最近 20 笔交易) 计算的盈利次数占比。
*   **盈亏比 (R/R Ratio)**: 实际结算的 `AvgProfit / AvgLoss`。
*   **期望值 (Expectancy)**: `(WinRate * AvgProfit) - (LossRate * AvgLoss)`。
    *   **硬性要求**: 任何处于激活状态的策略，其 50 笔交易后的滚动期望值必须为正数值。

### 7.2 动态风险降级 (Adaptive Risk Scaling)
*   **减半模式 (Half-Size Mode)**:
    *   触发条件: 最近 10 笔交易胜率 < 30% 或 期望值跌破手续费成本。
    *   动作: 强制将该策略的 `PositionSize` 压缩至原定的 50%。
*   **冷静期 (Cooling-off Period)**:
    *   触发条件: 触发 L1 级连续亏损熔断 (e.g., 5 连损)。
    *   动作: 该策略强制停止 24 小时，需人工检视行情匹配度。

### 7.3 胜率对止损的反馈机制
*   **止损收紧**: 若胜率高但单笔回撤大，系统会自动建议或强制收紧 `StopLossPrice` 的 ATR 倍数。
*   **保本触发**: 当盈利达到 RRR 1.2:1 后，系统自动开启“保本触发器”，将止损位移至 `EntryPrice`，确保该笔交易不再产生本金损失。

---

## 8. 风控配置模型

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

