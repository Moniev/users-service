from __future__ import annotations
from app.config.database import Base
from datetime import datetime, timedelta 
import uuid
from loguru import logger
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, String, select, Select
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession 
from typing import Any, Dict, Optional, TYPE_CHECKING


if TYPE_CHECKING:
    from .user import User


class ActivationCode(Base):
    __bind_key__ = "ActivationCode"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    code: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    used: Mapped[bool] = mapped_column(Boolean, default=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    expires_at: Mapped[datetime] = mapped_column(DateTime, default=lambda: datetime.now() + timedelta(hours=24))

    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"), unique=True, nullable=False)
    user: Mapped["User"] = relationship(back_populates="activation_code")
    __tablename__ = "ActivationCode"


    def __init__(self, code: str, user_id: int):
        self.code: str = code
        self.user_id: int = user_id
    
    
    def to_dict(self) -> Dict[str, Any]:
        pass
    
    
    def from_dict(self) -> Dict[str, Any]:
        pass
    
    @classmethod
    async def get_activation_code_by_id(cls, id: int, async_session: async_sessionmaker[AsyncSession]) -> Optional['ActivationCode']:
        async with async_session() as session:
            statement: Select = select(cls).where(cls.id == id)
            result = await session.execute(statement)
            return result.scalars().first()
    
    @classmethod
    async def get_activation_code_by_code(cls, code: str, async_session: async_sessionmaker[AsyncSession]) -> Optional['ActivationCode']:
        async with async_session() as session:
            statement: Select = select(cls).where(cls.code == code)
            result = await session.execute(statement)
            return result.scalars().first()
    
    @classmethod
    def create_activation_code(cls, user_id: int) -> 'ActivationCode':
        code_value = str(uuid.uuid4()) 
        new_code = cls(code=code_value, user_id=user_id) 
        logger.info(f"Generated new ActivationCode object in memory for user ID {user_id}: {new_code.code[:8]}...")
        return new_code