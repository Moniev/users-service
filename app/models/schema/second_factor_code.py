from __future__ import annotations
from app.config.database import Base
from datetime import datetime
from pydantic import BaseModel
import uuid 
from loguru import logger
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import Optional
from .user import User

class SecondFactorCode(Base):
    __bind_key__ = "SecondFactorCode"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    code: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    used: Mapped[bool] = mapped_column(Boolean, default=False)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="second_factor_code")
    
    __tablename__ = "SecondFactorCode"
    
    def __init__(self, code: str, user_id: int): 
        self.code: str = code
        self.user_id: int = user_id
    
    @staticmethod 
    def create_second_factor_code() -> str: 
        code_value = str(uuid.uuid4())
        logger.info(f"Generated raw SecondFactorCode string: {code_value[:8]}...")
        return code_value

