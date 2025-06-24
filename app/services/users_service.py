from app.config.redis import get_redis_client
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from redis.asyncio.client import Redis
from aiokafka import AIOKafkaProducer
from loguru import logger
from typing import Dict

class UsersService:
    pass
        