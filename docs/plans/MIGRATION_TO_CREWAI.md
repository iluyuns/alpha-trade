# Migration Plan: From LangGraph to CrewAI

## 1. 迁移背景 (Rationale)
为了实现更高程度的自动化任务处理（多源调研、复杂逻辑推演、策略自动生成），系统决定将 AI 决策层框架从原有的 `LangGraph` (状态机模式) 迁移至 `CrewAI` (多智能体协作模式)。

## 2. 核心架构变更 (Architectural Changes)

### 2.1 逻辑抽象层
*   **原方案 (LangGraph)**: 预定义节点流 (Crawler -> Analyzer -> Decider)。
*   **新方案 (CrewAI)**: 基于角色的协作 (Roles: Scraper, Analyst, Strategist)。

### 2.2 数据流 (Data Flow)
`MARKET.NEWS (NATS)` -> `Crew.kickoff()` -> `Multi-Agent Collaboration` -> `AI.DECISION (NATS)`。

## 3. 实施步骤 (Execution Steps)

### Phase 1: 环境准备 (Environment)
- [ ] 更新 `requirements.txt`：移除 `langgraph`，添加 `crewai`。
- [ ] 安装依赖并验证 Gemini 连通性。

### Phase 2: Agent 与 Task 定义 (Definition)
- [ ] **Scraper Agent**: 封装 Jina Reader 工具。
- [ ] **Analyst Agent**: 注入加密货币黑话与市场分析 Prompt。
- [ ] **Strategist Agent**: 负责聚合结论并格式化为 JSON。
- [ ] 定义对应的 **Tasks**，确保输出格式符合 Go 核心层的接口规范。

### Phase 3: 核心逻辑重构 (Refactoring)
- [ ] 重写 `ai-agent/agent.py`：移除 `StateGraph` 相关代码，改用 `Crew` 类。
- [ ] 更新 `ai-agent/main.py`：在 NATS 消息回调中触发 Crew 执行。

### Phase 4: 测试与验证 (Validation)
- [ ] 单元测试：各 Agent 的 Prompt 响应率。
- [ ] 集成测试：端到端延迟测试 (End-to-End Latency)。
- [ ] 鲁棒性测试：心跳机制在 Crew 运行期间的稳定性。

## 4. 进度跟踪
- **2025-01-01**: 启动迁移计划，完成文档更新。

