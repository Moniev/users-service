from __future__ import annotations
from app.config.database import Base
import asyncio
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError
from datetime import datetime
from loguru import logger
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String, Table, Column, select, Select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import List, Optional, Tuple, TYPE_CHECKING

if TYPE_CHECKING:
    from .user_details import UserDetails
    from .user_device import UserDevice
    from .user_role import UserRole
    from .user_action import UserAction
    from .user_settings import UserSettings
    from .user_role import UserRole
    from .reset_code import ResetCode
    from .activation_code import ActivationCode
    from .second_factor_code import SecondFactorCode
    from .verification_code import VerificationCode

password_hasher: PasswordHasher = PasswordHasher(
    time_cost=32,
    memory_cost=256,
    parallelism=2,
    hash_len=256,
    salt_len=16,
    encoding="utf-8"
)

user_action_association_table: Table = Table(
    "user_action_association",
    Base.metadata,
    Column("user_id", ForeignKey("User.id"), primary_key=True),
    Column("user_action_id", ForeignKey("UserAction.id"), primary_key=True),
)

user_role_association_table: Table = Table(
    "user_role_association",
    Base.metadata,
    Column("user_id", ForeignKey("User.id"), primary_key=True),
    Column("user_role_id", ForeignKey("UserRole.id"), primary_key=True),
)


class User(Base):    
    __bind_key__ = "User"
    id: Mapped[int] = mapped_column(Integer, autoincrement=True, primary_key=True, nullable=False, unique=True)
    uuid: Mapped[str] = mapped_column(String(64), nullable=True, unique=True)
    mail: Mapped[str] = mapped_column(String, nullable=False, unique=True)
    phone: Mapped[str] = mapped_column(String, nullable=True, unique=True)
    password: Mapped[str] = mapped_column(String, nullable=False, unique=False)
    active: Mapped[bool] = mapped_column(Boolean, default=False)
    verified: Mapped[bool] = mapped_column(Boolean, default=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    
    details: Mapped[Optional["UserDetails"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    devices: Mapped[List["UserDevice"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    user_roles: Mapped[List["UserRole"]] = relationship(secondary=user_role_association_table, back_populates="users")
    user_actions: Mapped[List["UserAction"]] = relationship(secondary=user_action_association_table, back_populates="users")
    settings: Mapped[Optional["UserSettings"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    reset_code: Mapped[Optional["ResetCode"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    activation_code: Mapped[Optional["ActivationCode"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    second_factor_code: Mapped[Optional["SecondFactorCode"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    verification_code: Mapped[Optional["VerificationCode"]] = relationship(back_populates="user", cascade="all, delete-orphan")
    
    __tablename__ = "User"
    
    
    def __init__(self, mail: str, phone: str, password: str, uuid: str=None, active: bool = False, verified: bool = False):
        self.mail: str = mail
        self.phone: str = phone
        self.password: str = password_hasher.hash(password)
        self.uuid: str = uuid
        self.active: bool = active
        self.verified: bool = verified

    def __repr__(self):
        pass

    @classmethod
    async def get_user_by_id(cls, async_session: async_sessionmaker[AsyncSession], id: int) -> User:
        async with async_session() as session:
            statement: Select = select(User).where(User.id == id)
            result = await session.execute(statement)
            await session.close()
            return result.scalars().first()
    
    @classmethod
    async def get_user_by_mail(cls, async_session: async_sessionmaker[AsyncSession], mail: str) -> User:
        async with async_session() as session:
            statement: Select = select(User).where(User.mail == mail)
            result = await session.execute(statement)
            await session.close()
            return result.scalars().first()
    
    @classmethod
    async def get_user_by_activation_code(cls, async_session: async_sessionmaker[AsyncSession], code: str) -> User:
        from app.models.schema import ActivationCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.activation_code) 
                .where(ActivationCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            await session.close()
            return result.scalars().first()
    
    @classmethod
    async def get_user_by_verification_code(cls, async_session: async_sessionmaker[AsyncSession], code: str) -> User:
        from app.models.schema import VerificationCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.verification_code) 
                .where(VerificationCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            await session.close()
            return result.scalars().first()
    
    @classmethod
    async def get_user_by_second_factor_code(cls, async_session: async_sessionmaker[AsyncSession], code: str) -> User:
        from app.models.schema import SecondFactorCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.second_factor_code) 
                .where(SecondFactorCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            await session.close()
            return result.scalars().first()
        
    @classmethod
    async def get_user_by_reset_code(cls, async_session: async_sessionmaker[AsyncSession], code: str) -> User:
        from app.models.schema import ResetCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.reset_code) 
                .where(ResetCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            await session.close()
            return result.scalars().first()
    
    @classmethod
    async def verify_password(cls, async_session: async_sessionmaker[AsyncSession], mail: str, password: str) -> User:
        user: Optional[User] = await cls.get_user_by_mail(async_session, mail)

        if not user:
            return None

        try:
            await asyncio.to_thread(password_hasher.verify, user.password, password)
            
            logger.info(f"Successfully verified user with ID: {user.id}")
            return user

        except VerifyMismatchError:
            logger.info(f"Provided wrong password for user with ID: {user.id}")
            return None
    
    @classmethod
    async def register(cls, async_session: async_sessionmaker[AsyncSession], mail: str, password: str) -> Optional[Tuple[User, ActivationCode]]:
        async with async_session() as session:
            existing_user = await cls.get_user_by_mail(async_session=async_session, mail=mail)
            if existing_user:
                logger.warning(f"Mail already exists is database")
                return None

            try:
                user = User(mail, None, password)
                activation_code = ActivationCode.create_activation_code() 
                user.activation_code = activation_code
                session.add(user)

                await session.commit()

                await session.refresh(user)
                await session.refresh(activation_code)
                
                logger.info(f"Succesfully created user with ID: {user.id}")
                return user, activation_code

            except IntegrityError as e:
                await session.rollback()
                logger.error(f"{e}")
                return None

    async def activate_account(self, async_session: async_sessionmaker[AsyncSession], activation_code: ActivationCode) -> Optional['User']:
        from app.models.schema import ActivationCode
        
        if self.active:
            logger.info("User is already activated")
            return None
        
        if not self.activation_code or self.activation_code.code != activation_code.code:
            logger.error(f"Failed to activate user with ID: {self.id}. Failed to find activation code.")
            return None

        async with async_session() as session:
            try:
                user_in_session: User = await session.merge(self)
                code_in_session: ActivationCode = await session.merge(self.activation_code)

                user_in_session.active = True
                await session.delete(code_in_session)

                await session.commit()

                logger.info(f"Successfully activated user with ID: {user_in_session.id}")
                return user_in_session

            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Integrity error for user with ID: {self.id}. {e}")
                return None
            except Exception as e:
                await session.rollback()
                logger.error(f"Unknown error for user with ID: {self.id}: {e}")
                return None
        
        logger.error(f"Failed to activate user with ID: {self.id}")
                
    
    async def create_verification_code():
        pass
    
    async def verify_account():
        pass
    
    async def add_user_role():
        pass
    
    async def revoke_user_role():
        pass
    
    async def modify_user_settings():
        pass
    
    async def modify_user_details():
        pass
    
    async def add_user_action():
        pass