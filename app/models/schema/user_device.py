from __future__ import annotations
from ...config.database import Base
from .user import User
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func



class UserDevice(Base):
    __bind_key__ = "UserDevice"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    device_id: Mapped[str] = mapped_column(String, unique=True)
    device_type: Mapped[str] = mapped_column(String(50))
    last_login: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="devices")
    
    __tablename__ = "UserDevice"
    
    def __repr__(self):
        return f"<UserDevice(id={self.id}, type='{self.device_type}', user_id={self.user_id})>"