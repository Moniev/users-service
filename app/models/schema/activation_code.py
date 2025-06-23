from __future__ import annotations
from ...config.database import Base
from datetime import datetime
from loguru import logger
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from .user import User


class ActivationCode(Base):
    __bind_key__ = "ActivationCode"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    code: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"), nullable=False)
    user: Mapped["User"] = relationship(back_populates="activation_code")
    
    __tablename__ = "ActivationCode"
    
    def __init__(self, code: str, user_id: int):
        self.code: str = code
        self.user_id: int = user_id
    
    @classmethod
    def get_activation_code_by_id(cls, id: int) -> ActivationCode:
        pass
    
    @classmethod
    def get_activation_code_by_code(cls, code: str) -> ActivationCode:
        pass
    
    @classmethod
    def create_activation_code(cls) -> ActivationCode:
        pass