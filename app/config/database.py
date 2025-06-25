import os
import ssl
from loguru import logger
from typing import AsyncGenerator, Optional
from sqlalchemy.ext.asyncio import (
    create_async_engine,
    async_sessionmaker,
    AsyncSession,
    AsyncEngine,
)
from sqlalchemy.orm import DeclarativeBase
from .settings import settings


class Base(DeclarativeBase):
    pass


def get_ssl_context() -> Optional[ssl.SSLContext]:
    if os.environ.get("TEST_MODE") == "True":
        return None  

    if not all([settings.DB_SSL_CA_PATH, settings.DB_SSL_CERT_PATH, settings.DB_SSL_KEY_PATH]):
        logger.warning("One or more DB SSL certificate paths are not set. Proceeding without SSL.")
        return None
    try:
        context = ssl.create_default_context(cafile=settings.DB_SSL_CA_PATH)
        context.load_cert_chain(certfile=settings.DB_SSL_CERT_PATH, keyfile=settings.DB_SSL_KEY_PATH)
        logger.info("Database SSL context created successfully.")
        return context
    except Exception as e:
        logger.error(f"Error creating database SSL context: {e}")
        raise


if os.environ.get("TEST_MODE") == "True":
    logger.warning("TEST_MODE is active. Using in-memory SQLite database for tests.")
    DATABASE_URI = "sqlite+aiosqlite:///:memory:"
    connect_args = {}
else:
    logger.info(f"Database URI configured for host: {settings.DB_HOST}")
    DATABASE_URI = settings.DATABASE_URI
    ssl_context = get_ssl_context()
    connect_args = {"ssl": ssl_context} if ssl_context else {}

engine: AsyncEngine = create_async_engine(
    DATABASE_URI,
    echo=False,
    connect_args=connect_args
)

async_session_factory: async_sessionmaker[AsyncSession] = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

async def create_db_and_tables():
    logger.info("Initializing database schema...")
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    logger.info("Database schema initialized successfully.")

async def get_async_session() -> AsyncGenerator[AsyncSession, None]:
    async with async_session_factory() as session:
        yield session
        
def async_session_loader(connection: AsyncEngine) -> async_sessionmaker[AsyncSession]:
    if connection:
        session = async_sessionmaker(
            bind=connection,
            expire_on_commit=False
        )
        return session
    
def get_session_factory() -> async_sessionmaker[AsyncSession]:
    return async_session_factory
