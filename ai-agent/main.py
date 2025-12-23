import asyncio
import json
import os
import nats
from nats.js.errors import NotFoundError
from agent import app
from dotenv import load_dotenv

load_dotenv()

NATS_URL = os.getenv("NATS_URL", "nats://localhost:4222")
HEARTBEAT_INTERVAL = 5 # 秒

async def send_heartbeat(nc):
    """持续发送心跳消息"""
    while True:
        try:
            heartbeat_data = {"service": "ai-agent", "status": "alive", "timestamp": asyncio.get_event_loop().time()}
            await nc.publish("AI.HEARTBEAT", json.dumps(heartbeat_data).encode())
            await asyncio.sleep(HEARTBEAT_INTERVAL)
        except Exception as e:
            print(f"Heartbeat error: {e}")
            await asyncio.sleep(1)

async def main():
    # 连接 NATS
    nc = await nats.connect(NATS_URL)
    js = nc.jetstream()

    # 确保 Stream 存在 (如果不存在则创建)
    # 在 JetStream 中，必须先定义 Stream 才能订阅特定的 Subject
    try:
        await js.add_stream(name="ALPHATRADE", subjects=["MARKET.*", "AI.*"])
        print("Stream 'ALPHATRADE' created or already exists.")
    except Exception as e:
        print(f"Stream info: {e}")

    print(f"Connected to NATS at {NATS_URL}")

    # 启动心跳协程
    asyncio.create_task(send_heartbeat(nc))

    # 订阅新闻 Topic (示例使用 JetStream)
    sub = await js.subscribe("MARKET.NEWS", durable="ai_agent_worker")

    async for msg in sub.messages:
        try:
            data = json.loads(msg.data.decode())
            news_text = data.get("text", "")
            
            print(f"Processing news: {news_text[:50]}...")

            # 调用 LangGraph 进行推理
            result = await asyncio.to_thread(app.invoke, {"input": news_text, "logs": []})

            # 将结果发回 MQ
            decision_data = {
                "sentiment": result["sentiment_score"],
                "decision": result["decision"],
                "confidence": result["confidence"],
                "source": "Gemini-3-LangGraph"
            }
            await nc.publish("AI.DECISION", json.dumps(decision_data).encode())
            
            # 手动 ACK
            await msg.ack()
            print(f"Decision sent: {result['decision']}")

        except Exception as e:
            print(f"Error processing message: {e}")

if __name__ == "__main__":
    asyncio.run(main())

