from loguru import logger
import redis.asyncio as redis
from redis.asyncio.client import Redis
from redis.asyncio.connection import ConnectionPool
from typing import Any, Dict
from .settings import settings

logger.info(f"Preparing Redis connection for {settings.REDIS_HOST}:{settings.REDIS_PORT}")

pool_kwargs = {
    "max_connections": 10,
    "decode_responses": True
}

if settings.REDIS_SSL_CA_PATH:
    logger.info("Configuring Redis connection with TLS/SSL using from_url...")
    pool_kwargs.update({
        "ssl_cert_reqs": "required",
        "ssl_ca_certs": settings.REDIS_SSL_CA_PATH,
        "ssl_certfile": settings.REDIS_SSL_CERT_PATH,
        "ssl_keyfile": settings.REDIS_SSL_KEY_PATH,
})
else:
    logger.warning("Redis SSL certificates not configured. Connection will not use TLS.")

logger.info(f"Preparing Redis connection for {settings.REDIS_HOST}:{settings.REDIS_PORT}")

pool: ConnectionPool = redis.ConnectionPool.from_url(
    settings.REDIS_URI, 
    **pool_kwargs
)
redis_client: Redis = redis.Redis(connection_pool=pool)


async def get_redis_client() -> Redis:
    return redis_client

async def close_redis_connection() -> None:
    logger.info("Closing Redis connection pool...")
    await redis_client.close()
    await redis_client.connection_pool.disconnect()
    logger.info("Redis connection pool closed.")