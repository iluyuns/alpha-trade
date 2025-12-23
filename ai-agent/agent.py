import os
import asyncio
import httpx
from typing import Annotated, TypedDict
from langgraph.graph import StateGraph, END
from langchain_google_genai import ChatGoogleGenerativeAI
from dotenv import load_dotenv

load_dotenv()

# 定义 Agent 状态
class AgentState(TypedDict):
    input: str        # 原始新闻标题/简述
    url: str          # 新闻链接
    full_content: str # 抓取到的完整网页内容
    sentiment_score: float
    decision: str
    confidence: float
    logs: list[str]

# 初始化 Gemini 3
def get_model():
    return ChatGoogleGenerativeAI(
        model="gemini-1.5-flash",
        google_api_key=os.getenv("GOOGLE_API_KEY"),
        temperature=0.1
    )

# 节点 1: 网页内容抓取 (Jina Reader)
def fetch_web_content(state: AgentState):
    url = state.get("url")
    if not url:
        return {"full_content": state["input"], "logs": state["logs"] + ["No URL provided, using input text"]}
    
    try:
        jina_url = f"https://r.jina.ai/{url}"
        headers = {"X-Return-Format": "markdown"}
        # 由于 LangGraph 节点目前主要以同步方式调用，这里使用 sync 模式或 asyncio.run
        with httpx.Client() as client:
            response = client.get(jina_url, headers=headers, timeout=20.0)
            if response.status_code == 200:
                return {
                    "full_content": response.text,
                    "logs": state["logs"] + [f"Web content fetched from {url} via Jina"]
                }
    except Exception as e:
        print(f"Jina fetch error: {e}")
    
    return {"full_content": state["input"], "logs": state["logs"] + ["Fetch failed, using input text"]}

# 节点 2: 情绪分析
def analyze_sentiment(state: AgentState):
    model = get_model()
    # 使用完整内容进行分析
    content = state.get("full_content", state["input"])
    prompt = f"分析以下加密货币新闻的完整内容或摘要，给出 -1 到 1 之间的情绪得分：\n\n{content}"
    response = model.invoke(prompt)
    
    return {
        "sentiment_score": 0.0, # 实际需解析 response
        "logs": state["logs"] + ["Sentiment analysis completed"]
    }

# 节点 3: 决策逻辑
def make_decision(state: AgentState):
    score = state["sentiment_score"]
    decision = "NEUTRAL"
    if score > 0.5:
        decision = "LONG_ONLY"
    elif score < -0.5:
        decision = "HALT_TRADING"
    
    return {
        "decision": decision,
        "confidence": 0.85,
        "logs": state["logs"] + [f"Decision made: {decision}"]
    }

# 构建图
def create_graph():
    workflow = StateGraph(AgentState)

    workflow.add_node("crawler", fetch_web_content)
    workflow.add_node("analyzer", analyze_sentiment)
    workflow.add_node("decider", make_decision)

    workflow.set_entry_point("crawler")
    workflow.add_edge("crawler", "analyzer")
    workflow.add_edge("analyzer", "decider")
    workflow.add_edge("decider", END)

    return workflow.compile()

# 导出应用
app = create_graph()
