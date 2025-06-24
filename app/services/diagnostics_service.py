from app.config.redis import get_redis_client
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from redis.asyncio.client import Redis
from aiokafka import AIOKafkaProducer
from loguru import logger
from typing import Dict

class DiagnosticsService:    
    def get_liveness_status(self) -> Dict[str, str]:
        return {"status": "ok"}

    async def get_readiness_status(
        self,
        session: AsyncSession,
        redis: Redis,
        kafka_producer: AIOKafkaProducer
    ) -> Dict[str, str]:
      
        try:
            await session.execute(text("SELECT 1"))
            db_status: str = "ok"
        except Exception as e:
            logger.error(f"Readiness check failed: Database connection error - {e}")
            db_status = "error"
            
        try:
            await redis.ping()
            redis_status = "ok"
        except Exception as e:
            logger.error(f"Readiness check failed: Redis connection error - {e}")
            redis_status = "error"

        
        kafka_status = "ok"
        if not kafka_producer or not await kafka_producer.bootstrap_connected():
            kafka_status = "error"
            logger.error("Readiness check failed: Kafka producer not connected.")
            
        return {
            "database": db_status,
            "redis": redis_status,
            "kafka": kafka_status
        }