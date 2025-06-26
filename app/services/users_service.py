from __future__ import annotations
from aiokafka import AIOKafkaProducer
from app.config.redis import get_redis_client
from app.models.schema import User, UserRole, UserDetails, UserSettings
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from loguru import logger
from typing import Dict, List, Optional, Any


class UsersService:
    def __init__(self, session_factory: async_sessionmaker[AsyncSession], kafka_producer: AIOKafkaProducer = None):
        self.async_session: async_sessionmaker[AsyncSession] = session_factory
        self.kafka_producer: AIOKafkaProducer = kafka_producer


    async def get_all_users(self, skip: int = 0, limit: int = 100) -> List[User]:
        logger.info(f"Fetching all users with skip: {skip}, limit: {limit}")
        async with self.async_session() as session:
            stmt = select(User).where(User.deleted == False).offset(skip).limit(limit)
            result = await session.execute(stmt)
            return result.scalars().all()


    async def get_user_full_details(self, user_id: int) -> Optional[User]:
        logger.info(f"Fetching full details for user ID: {user_id}")
        async with self.async_session() as session:
            stmt = select(User).where(User.id == user_id).options(
                selectinload(User.details),
                selectinload(User.devices),
                selectinload(User.user_roles),
                selectinload(User.user_actions),
                selectinload(User.settings)
            )
            result = await session.execute(stmt)
            return result.scalars().one_or_none()


    async def update_user_core_details(self, user_id: int, details_to_update: Dict[str, Any]) -> Optional[User]:
        logger.info(f"Updating core details for user ID: {user_id}")
        async with self.async_session() as session:
            user = await session.get(User, user_id)
            if not user:
                logger.error(f"Update failed: User with ID {user_id} not found.")
                return None
            
            allowed_fields = {"mail", "phone", "uuid"}
            for key, value in details_to_update.items():
                if key in allowed_fields:
                    setattr(user, key, value)
            
            try:
                await session.commit()
                await session.refresh(user)
                logger.success(f"Successfully updated core details for user ID: {user_id}")
                return user
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to update user ID {user_id}: {e}")
                return None


    async def set_user_status(self, user_id: int, active: bool = None, verified: bool = None) -> Optional[User]:
        logger.info(f"Setting status for user ID: {user_id} (active={active}, verified={verified})")
        async with self.async_session() as session:
            user = await session.get(User, user_id)
            if not user:
                logger.error(f"Set status failed: User with ID {user_id} not found.")
                return None
            
            if active is not None:
                user.active = active
            if verified is not None:
                user.verified = verified
            
            await session.commit()
            await session.refresh(user)
            return user


    async def blacklist_user(self, user_id: int, is_blacklisted: bool) -> Optional[User]:
        status = "blacklisting" if is_blacklisted else "un-blacklisting"
        logger.info(f"Attempting to {status} user ID: {user_id}")
        async with self.async_session() as session:
            user = await session.get(User, user_id)
            if not user:
                logger.error(f"Blacklist operation failed: User with ID {user_id} not found.")
                return None
            
            user.black_listed = is_blacklisted
            await session.commit()
            await session.refresh(user)
            logger.success(f"Successfully updated blacklist status for user ID: {user_id}")
            return user


    async def soft_delete_user(self, user_id: int, is_deleted: bool) -> Optional[User]:
        status = "soft-deleting" if is_deleted else "restoring"
        logger.info(f"Attempting to {status} user ID: {user_id}")
        async with self.async_session() as session:
            user = await session.get(User, user_id)
            if not user:
                logger.error(f"Soft-delete operation failed: User with ID {user_id} not found.")
                return None
            
            user.deleted = is_deleted
            await session.commit()
            await session.refresh(user)
            logger.success(f"Successfully updated deleted status for user ID: {user_id}")
            return user
        