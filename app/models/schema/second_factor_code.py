from __future__ import annotations
from ...config.database import Base
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from .user import User

class SecondFactorCode(Base):
    __bind_key__ = "SecondFactorCode"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    code: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="second_factor_code")
    
    __tablename__ = "SecondFactorCode"