from __future__ import annotations
from app.config.database import Base
from datetime import datetime
from pydantic import BaseModel
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import Any, Dict, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from .user import User

class VerificationCode(Base):
    __bind_key__ = "VerificationCode"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    code: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="verification_code")
    __tablename__ = "VerificationCode"
    
    
    def to_dict(self) -> Dict[str, Any]:
        pass
    
    
    def from_dict(self) -> Dict[str, Any]:
        pass
    
    @classmethod
    def create_verification_code(cls, user: User) -> Optional['VerificationCode']:
        pass
    
    