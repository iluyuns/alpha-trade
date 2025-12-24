# Alpha-Trade AI Agent (Phase 1)

基于 LangGraph 和 Gemini 1.5 Flash 的交易辅助决策智能体，负责非结构化信息的量化与逻辑推演。

## 1. 核心职责边界 (Architecture Boundaries)

为了保证低延迟交易与高逻辑深度的平衡，系统采用了 **“快慢路径分离”** 的设计：

### 1.1 Go 核心侧 (The "Safety & Speed" Layer)
- **News Gateway (数据采集)**: 
    - 并发轮询 RSS (CoinDesk, CoinTelegraph)、交易所公告 API、以及通过 RSSHub 监控 X (Twitter) 关键账户。
    - **职责**: 负责极速获取原始数据并标准化，通过 NATS `MARKET.NEWS` 广播。
- **Heartbeat Monitor (安全卫兵)**: 
    - 实时监听 `AI.HEARTBEAT`。
    - **逻辑**: 若 30s 无心跳，自动触发 **Fail-safe 模式**（禁止所有新开仓，仅允许止损平仓）。
- **Risk Integration (指令执行)**: 
    - 订阅 `AI.DECISION`，将 AI 建议转化为风控参数（如动态调整杠杆上限、方向偏好控制）。

### 1.2 AI-Agent 侧 (The "Reasoning" Layer)
- **Content Expansion (全量抓取)**: 
    - **主选**: Jina Reader (`r.jina.ai`) 将原始 URL 转换为 LLM 易读的 Markdown。
    - **备选**: 集成 `Firecrawl` 或本地 `Trafilatura` 库，用于绕过复杂反爬并实现本地化的 HTML 降噪。
- **Cognitive Analysis (深度推演)**: 
    - **隐喻与黑话识别**: 结合内置 **"Crypto-Knowledge-Base"** 识别 Musk 发图、CZ 推文背后的币种关联。
    - **广告与噪音过滤**: 自动识别并剔除文中赞助商链接、SEO 堆砌及无关社交引导，确保 AI 仅分析核心事实。
    - **多源交叉验证**: 若 5min 内只有单一源报道重大新闻，标记为 Low Confidence。
- **Knowledge Base (知识库构建)**:
    - **Layer 1 (Entity Map)**: 常驻内存的别名表（如 `大饼 -> BTC`, `V神 -> Vitalik`），详见 `assets/knowledge.json`。
    - **Layer 2 (Historical Events)**: 重大历史事件库（如 `FTX 崩溃`, `ETF 审批`），包含影响权重与模式，详见 `assets/major_events.json`。
    - **Layer 3 (Context RAG)**: (中长期) 接入向量数据库，存储项目白皮书、历史极端行情复盘资料。
- **结构化输出**: 将 Gemini 的推理转化为精确 JSON（Score, Decision, Confidence）。

## 2. 待开发清单与难度评估

### 2.1 Go 侧待办 (Medium Difficulty)
- [ ] **NewsGateway 多源聚合**: 接入 RSSHub 和 Binance 公告流。
- [ ] **Fail-safe 状态机**: 在风控模块中实现心跳监测与自动降级逻辑。
- [ ] **NATS JetStream 调优**: 确保消息的“至少一次投递”与 Ack 机制。

### 2.2 AI-Agent 侧待办 (High Reasoning Difficulty)
- [ ] **结构化解析器**: 修复目前 `sentiment_score` 硬编码问题，实现 Gemini JSON Mode 解析。
- [ ] **隐喻知识库 (Prompt Engineering)**: 构建包含加密货币领域黑话、关联关系的 System Prompt。
- [ ] **长文本推理**: 优化 Jina 抓取后的文本分片处理，防止超长文本浪费 Token。

## 3. 开发路线图 (3-5 Days Plan)
- **D1 (数据源接入)**: Go 侧实现新闻网关。
- **D2 (逻辑推演)**: AI 侧完善 Prompt 和结构化解析。
- **D3 (安全兜底)**: 实现 Go 侧心跳监测与 Fail-safe。
- **D4 (全链路联调)**: 测试“推特发文 -> AI 分析 -> 风控拦截”的完整链路。

## 4. 快速开始

### 4.1 环境准备
确保已安装 Python 3.10+。

### 4.2 安装依赖
建议在虚拟环境中安装：
```bash
# 创建虚拟环境
python -m venv venv

# 激活虚拟环境 (Windows Git Bash)
source venv/Scripts/activate

# 安装依赖
pip install -r requirements.txt
```

### 4.3 配置环境变量
在 `ai-agent/` 目录下创建 `.env` 文件：
```env
GOOGLE_API_KEY=你的Gemini_API_Key
NATS_URL=nats://localhost:4222
```

### 4.4 运行
```bash
python main.py
```

## 5. 消息协议 (NATS)

| Topic | 模式 | 描述 |
| :--- | :--- | :--- |
| `MARKET.NEWS` | Subscribe (JS) | 接收原始新闻文本 (JSON) |
| `AI.DECISION` | Publish | 发送 AI 决策建议 (Score/Decision) |
| `AI.HEARTBEAT` | Publish | 发送心跳信号 (每 5s 一次) |

