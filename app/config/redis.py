from app.config.settings import Settings 
from loguru import logger
import redis.asyncio as redis
from redis.asyncio.client import Redis
from redis.asyncio.connection import ConnectionPool
from typing import Any, Dict, Optional


redis_client_instance: Optional[Redis] = None


async def get_redis_client(settings: Settings) -> Redis: 
    global redis_client_instance
    if redis_client_instance is not None:
        logger.info("Returning existing Redis client instance.")
        return redis_client_instance

    logger.info(f"Initializing Redis client for {settings.REDIS_HOST}:{settings.REDIS_PORT} using PROVIDED settings.")

    pool_kwargs = {
        "max_connections": 10,
        "decode_responses": True
    }

    if settings.REDIS_SSL_CA_PATH: 
        logger.info("Configuring Redis connection with TLS/SSL...")
        pool_kwargs.update({
            "ssl_cert_reqs": "required",
            "ssl_ca_certs": settings.REDIS_SSL_CA_PATH,
            "ssl_certfile": settings.REDIS_SSL_CERT_PATH,
            "ssl_keyfile": settings.REDIS_SSL_KEY_PATH,
        })
    else:
        logger.warning("Redis SSL certificates not configured. Connection will not use TLS.")

    try:
        pool: ConnectionPool = redis.ConnectionPool.from_url(
            settings.REDIS_URI, 
            **pool_kwargs
        )
        redis_client_instance = redis.Redis(connection_pool=pool)
        await redis_client_instance.ping() 
        logger.info("Redis client initialized and connected successfully.")
    except Exception as e:
        logger.error(f"Failed to initialize Redis client: {e}. Connection params: {settings.REDIS_URI}")
        raise RuntimeError(f"Could not connect to Redis: {e}")

    return redis_client_instance


async def close_redis_connection() -> None:
    global redis_client_instance 
    if redis_client_instance: 
        logger.info("Closing Redis connection pool...")
        await redis_client_instance.close() 
        await redis_client_instance.connection_pool.disconnect()
        redis_client_instance = None 
        logger.info("Redis connection pool closed.")