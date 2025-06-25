from __future__ import annotations
from app.config.database import Base
from sqlalchemy import Boolean, ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .user import User


class UserDetails(Base):
    __bind_key__ = "UserDetails"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    first_name: Mapped[str] = mapped_column(String(50))
    last_name: Mapped[str] = mapped_column(String(50))
    city: Mapped[str] = mapped_column(String(200))
    street: Mapped[str] = mapped_column(String(200))
    building: Mapped[str] = mapped_column(String(200))
    apartment: Mapped[str] = mapped_column(String(200))
    postal_code: Mapped[str] = mapped_column(String(6))
    deleted: Mapped[bool] = mapped_column(Boolean, default=False)
    
    user_id: Mapped[int] = mapped_column(ForeignKey("User.id"))
    user: Mapped["User"] = relationship(back_populates="details")
    
    __tablename__ = "UserDetails"