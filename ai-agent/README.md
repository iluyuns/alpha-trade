# Alpha-Trade AI Agent

基于 LangGraph 和 Gemini 3 的交易辅助决策智能体。

## 核心功能

- **情绪分析**: 利用 Gemini 3 (1.5 Flash/Pro) 对非结构化新闻、公告进行实时情绪打分。
- **自动化决策**: 基于 LangGraph 构建决策流，输出交易偏向（Long/Short）或熔断指令。
- **异步驱动**: 通过 NATS JetStream 与 Go 交易内核通信，实现解耦。
- **心跳监测**: 持续发送心跳，确保 Go 核心能感知 AI 层的存活状态。

## 快速开始

### 1. 环境准备
确保已安装 Python 3.10+。

### 2. 安装依赖
建议在虚拟环境中安装：
```bash
# 创建虚拟环境
python -m venv venv

# 激活虚拟环境 (Windows Git Bash)
source venv/Scripts/activate

# 安装依赖
pip install -r requirements.txt
```

### 3. 配置环境变量
在 `ai-agent/` 目录下创建 `.env` 文件：
```env
GOOGLE_API_KEY=你的Gemini_API_Key
NATS_URL=nats://localhost:4222
```

### 4. 运行
```bash
python main.py
```

## 消息协议 (NATS)

| Topic | 模式 | 描述 |
| :--- | :--- | :--- |
| `MARKET.NEWS` | Subscribe (JS) | 接收原始新闻文本 |
| `AI.DECISION` | Publish | 发送 AI 决策建议 (Score/Decision) |
| `AI.HEARTBEAT` | Publish | 发送心跳信号 (每 5s 一次) |

## 目录结构
- `agent.py`: LangGraph 工作流定义与节点逻辑。
- `main.py`: NATS 连接管理、消息循环与心跳实现。
- `requirements.txt`: 项目依赖项。

