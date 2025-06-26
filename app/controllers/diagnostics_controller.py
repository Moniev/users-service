from aiokafka import AIOKafkaProducer
from app.config.database import get_async_session
from app.config.redis import get_redis_client
from app.config.kafka import get_kafka_producer
from app.services.diagnostics_service import DiagnosticsService
from fastapi import APIRouter, Depends, HTTPException, status
from redis.asyncio.client import Redis
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Dict


router = APIRouter(
    prefix="/diagnostics",
    tags=["Diagnostics & Health Checks"]
)

diagnostics_service: DiagnosticsService = DiagnosticsService()


@router.get("/health", status_code=status.HTTP_200_OK)
async def health_check():
    return diagnostics_service.get_liveness_status()


@router.get("/readiness", status_code=status.HTTP_200_OK)
async def readiness_check(
    session: AsyncSession = Depends(get_async_session),
    redis: Redis = Depends(get_redis_client),
    kafka_producer: AIOKafkaProducer = Depends(get_kafka_producer)
):
    
    dependency_statuses: Dict[str, str] = await diagnostics_service.get_readiness_status(
        session=session,
        redis=redis,
        kafka_producer=kafka_producer
    )

    if any(status == "error" for status in dependency_statuses.values()):
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=dependency_statuses
        )

    return dependency_statuses