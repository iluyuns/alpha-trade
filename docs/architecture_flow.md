# 全链路驱动架构图 (Full-Stack Data Flow)

本文档通过 Mermaid 流程图展示系统从行情采集、AI 决策到订单执行的完整链路。
Phase 3 仅走 Go 内部同步路径；NATS 与 AI 相关链路为 Phase 4+ 计划。

```mermaid
graph TD
    %% ==================================================
    %% 外部数据源层 (External Data Sources)
    %% ==================================================
    subgraph "External Sources"
        MD["币安行情 WebSocket (Ticks/Candles)"]
        RSS["新闻 RSS / 交易所公告 API"]
        Twitter["Twitter/X via RSSHub (Social)"]
        OnChain["链上异动 (Whale Alert/RPC) - Future"]
        L2Depth["L2/L3 订单簿深度 (Microstructure)"]
    end

    %% ==================================================
    %% 接入层 (Gateway Layer)
    %% ==================================================
    subgraph "Go Gateways"
        GW_Trade["Trade Gateway (Go)"]
        GW_News["News Gateway (Go)"]
        GW_OnChain["On-chain Gateway (Go) - Future"]
    end

    MD -->|1. 高频行情推送| GW_Trade
    RSS -->|2. 原始新闻/公告| GW_News
    Twitter -->|3. 社交异动信号| GW_News
    OnChain -->|3.1 大额转账信号| GW_OnChain
    L2Depth -->|1.1 挂单墙/失衡数据| GW_Trade

    %% ==================================================
    %% 消息中枢 (Message Bus) - Phase 4+
    %% ==================================================
    subgraph "NATS JetStream (Async Hub - Phase 4+)"
        MQ{{"NATS 流转中心"}}
    end

    GW_Trade -->|4. 标准化数据| MQ
    GW_News -->|5. 标准化新闻 JSON| MQ
    GW_OnChain -->|5.1 链上事件 JSON| MQ

    %% ==================================================
    %% 决策大脑 (Decision Engines)
    %% ==================================================
    
    %% 路径 A: 极速计算路径 (肌肉)
    subgraph "Go Trading Kernel"
        Strategy["Strategy Engine (Go)"]
        Risk["Risk Manager (Go)"]
        OMS["Order Management (Go)"]
    end

    %% 路径 B: 智能理解路径 (大脑)
    subgraph "Python AI Logic"
        AI_Agent["Multi-Agent Crew (Python/CrewAI)"]
        LLM[("Gemini 3 (LLM)")]
    end

    %% AI 驱动流
    MQ -->|6. 订阅分析请求| AI_Agent
    AI_Agent <-->|"7. 多角色协作推理"| LLM
    AI_Agent -->|"8. 发送 AI 偏见指令 (Bias/Halt)"| MQ

    %% 行情驱动流
    MQ -->|9. 触发策略计算 (Phase 4+)| Strategy
    MQ -->|10. 注入 AI 偏见参数 (Phase 4+)| Strategy
    GW_Trade -->|9. 行情直驱 (Phase 3)| Strategy
    
    %% ==================================================
    %% 执行层 (Execution Layer)
    %% ==================================================
    Strategy -->|11. 产生交易信号| Risk
    MQ -->|12. 广播紧急熔断信号 (Phase 4+)| Risk
    
    Risk -->|13. 风险合规审核| OMS
    OMS -->|14. 路由订单请求| GW_Trade
    GW_Trade -->|15. 执行 API 调用| MD

    %% ==================================================
    %% 可观测性 (Observability) - 独立于业务主链路
    %% ==================================================
    subgraph "Monitoring (Grafana Cloud)"
        OTEL["OTEL Collector"]
        Strategy -.->|Span with TraceID| OTEL
        AI_Agent -.->|Span with TraceID| OTEL
        Risk -.->|Span with TraceID| OTEL
    end

    %% 样式美化
    style MQ fill:#f9f,stroke:#333,stroke-width:2px
    style AI_Agent fill:#bbf,stroke:#333,stroke-width:2px
    style LLM fill:#dfd,stroke:#333,stroke-width:1px
    style Strategy fill:#ffd,stroke:#333,stroke-width:2px
```
