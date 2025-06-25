from __future__ import annotations
from app.config.database import Base
from .user import user_role_association_table, User
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String, select, Select
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import List, Optional


class UserRole(Base):
    __bind_key__ = "UserRole"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    name: Mapped[str] = mapped_column(String, nullable=False, unique=False)
    deleted: Mapped[bool] = mapped_column(Boolean, default=False)
    
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    users: Mapped[List["User"]] = relationship(
        secondary=user_role_association_table, back_populates="user_roles"
    )
    
    __tablename__ = "UserRole"
    
    @classmethod
    async def get_user_role_by_user_id(cls, async_session: async_sessionmaker[AsyncSession], user_id: int) -> Optional['UserRole']:
         async with async_session() as session:
            statement: Select = (
                select(UserRole)
                .join(User) 
                .where(User.id == user_id) 
            )
            
            result = await session.execute(statement)
            
            return result.scalars().first()