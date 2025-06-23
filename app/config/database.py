from .settings import settings
import ssl
from loguru import logger
from typing import AsyncGenerator
from sqlalchemy.ext.asyncio import (
    create_async_engine,
    async_sessionmaker,
    AsyncSession,
    AsyncEngine,
)
from sqlalchemy.orm import DeclarativeBase


class Base(DeclarativeBase):
    pass

def get_ssl_context() -> ssl.SSLContext | None:
    if not all([settings.DB_SSL_CA_PATH, settings.DB_SSL_CERT_PATH, settings.DB_SSL_KEY_PATH]):
        return None
    try:
        context = ssl.create_default_context(cafile=settings.DB_SSL_CA_PATH)
        context.load_cert_chain(certfile=settings.DB_SSL_CERT_PATH, keyfile=settings.DB_SSL_KEY_PATH)
        return context
    except Exception as e:
        logger.error(f"Error creating SSL context: {e}")
        raise

logger.info(f"Database URI configured for host: {settings.DB_HOST}")

connect_args = {"ssl": get_ssl_context()} if get_ssl_context() else {}

engine: AsyncEngine = create_async_engine(
    settings.DATABASE_URI,
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
        
        logger.debug()
        return session
    
def get_session_factory() -> async_sessionmaker[AsyncSession]:
    return async_session_factory