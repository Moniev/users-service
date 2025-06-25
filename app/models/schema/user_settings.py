from __future__ import annotations
from app.config.database import Base
from .user import User
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func


class UserSettings(Base):
    __bind_key__ = "UserSettings"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    theme: Mapped[str] = mapped_column(String(20), default="light")
    notifications_enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    notifications_login_enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    notifications_personal_enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    notifications_tasks_enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    second_factor_enabled: Mapped[bool] = mapped_column(Boolean, default=False)
    deleted: Mapped[bool] = mapped_column(Boolean, default=False)
    
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="settings")
    
    __tablename__ = "UserSettings"
    
    def __repr__(self):
        return f"<UserSettings theme='{self.theme}' notifications={self.notifications_enabled} >"