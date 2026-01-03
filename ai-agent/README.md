# Alpha-Trade AI Agent (Phase 1)

基于 CrewAI 和 Gemini 1.5 Flash 的多智能体协作决策系统，负责非结构化信息的量化与逻辑推演。

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

### 1.2 AI-Agent 侧 (The "Reasoning" Layer - Powered by CrewAI)
系统引入 **Multi-Agent Collaboration** 模式，由多个专业 Agent 协作完成任务：
- **Scraper Agent (内容专家)**: 
    - 使用 Jina Reader 获取网页深度内容，识别并清洗广告噪音。
- **Analyst Agent (量化分析师)**: 
    - 结合加密货币黑话知识库，对内容进行深度推演和情绪建模。
- **Strategist Agent (决策官)**: 
    - 汇总各方信息，最终输出结构化的 `Score` 和 `Decision`。
- **Knowledge Base (知识库支持)**:
    - 别名表与历史重大事件库（`assets/` 目录下）。

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

