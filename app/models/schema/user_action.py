from __future__ import annotations
from ...config.database import Base
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import List, Optional
from .user import User, user_action_association_table

class UserAction(Base):
    __bind_key__ = "UserAction"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    type: Mapped[str] = mapped_column(String, nullable=False, unique=False)
    
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    users: Mapped[List["User"]] = relationship(
        secondary=user_action_association_table, back_populates="user_actions"
    )
    
    __tablename__ = "UserAction"
    
    