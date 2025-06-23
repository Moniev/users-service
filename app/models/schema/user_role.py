from __future__ import annotations
from ...config.database import Base
from .user import user_role_association_table, User
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import List, Optional


class UserRole(Base):
    __bind_key__ = "UserRole"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    name: Mapped[str] = mapped_column(String, nullable=False, unique=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    users: Mapped[List["User"]] = relationship(
        secondary=user_role_association_table, back_populates="user_roles"
    )
    
    __tablename__ = "UserRole"