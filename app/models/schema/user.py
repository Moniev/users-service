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
from typing import Any, Dict, List, Optional, Tuple, TYPE_CHECKING

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
        user: Optional['User'] = await cls.get_user_by_mail(async_session, mail)

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
        from .activation_code import ActivationCode
        async with async_session() as session:
            existing_user = await cls.get_user_by_mail(async_session=async_session, mail=mail)
            if existing_user:
                logger.warning(f"Mail '{mail}' already exists in database.")
                return None

            try:
                user = User(mail, None, password)
                activation_code = ActivationCode.create_activation_code() 
                user.activation_code = activation_code
                session.add(user)
                await session.commit()
                await session.refresh(user)
                await session.refresh(activation_code)
                logger.info(f"Successfully created user with ID: {user.id}")
                return user, activation_code
            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Error during registration: {e}")
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
        
    async def verify_account(self, async_session: async_sessionmaker[AsyncSession], verification_code: VerificationCode) -> Optional['User']:
        from app.models.schema import VerificationCode
        
        if self.verified:
            logger.info("User is already verified")
            return None
        
        if not self.verification_code or self.verification_code.code != verification_code.code:
            logger.error(f"Failed to verify user with ID: {self.id}. Failed to find activation code.")
            return None

        async with async_session() as session:
            try:
                user_in_session: User = await session.merge(self)
                code_in_session: VerificationCode = await session.merge(self.verification_code)

                user_in_session.verified = True
                await session.delete(code_in_session)

                await session.commit()

                logger.info(f"Successfully verified user with ID: {user_in_session.id}")
                return user_in_session

            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Integrity error for user with ID: {self.id}. {e}")
                return None
            except Exception as e:
                await session.rollback()
                logger.error(f"Unknown error for user with ID: {self.id}: {e}")
                return None
        
        logger.error(f"Failed to verify user with ID: {self.id}")
                
    async def create_verification_code(self, async_session: async_sessionmaker[AsyncSession]) -> Optional[VerificationCode]:
        from .verification_code import VerificationCode
        if self.verified:
            logger.info(f"User with ID {self.id} is already verified.")
            return None
            
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                if user_in_session.verification_code:
                    await session.delete(await session.merge(user_in_session.verification_code))
                
                new_code: VerificationCode = VerificationCode.create_verification_code()
                user_in_session.verification_code = new_code
                await session.commit()
                await session.refresh(new_code)
                logger.info(f"Created new verification code for user ID {self.id}")
                return new_code
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to create verification code for user ID {self.id}: {e}")
                return None

    async def add_user_role(self, async_session: async_sessionmaker[AsyncSession], role: UserRole) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                role_in_session = await session.merge(role)
                
                if role_in_session not in user_in_session.user_roles:
                    user_in_session.user_roles.append(role_in_session)
                    await session.commit()
                    logger.info(f"Added role '{role.name}' to user ID {self.id}")
                else:
                    logger.info(f"User ID {self.id} already has role '{role.name}'")
                
                return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to add role to user ID {self.id}: {e}")
                return None

    async def revoke_user_role(self, async_session: async_sessionmaker[AsyncSession], role: UserRole) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                role_to_revoke = next((r for r in user_in_session.user_roles if r.id == role.id), None)

                if role_to_revoke:
                    user_in_session.user_roles.remove(role_to_revoke)
                    await session.commit()
                    logger.info(f"Revoked role '{role.name}' from user ID {self.id}")
                else:
                    logger.warning(f"Role '{role.name}' not found for user ID {self.id}. Cannot revoke.")
                
                return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to revoke role from user ID {self.id}: {e}")
                return None

    async def modify_user_settings(self, async_session: async_sessionmaker[AsyncSession], settings_data: Dict[str, Any]) -> Optional[UserSettings]:
        from .user_settings import UserSettings
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                if not user_in_session.settings:
                    user_in_session.settings = UserSettings(user_id=user_in_session.id)
                
                for key, value in settings_data.items():
                    if hasattr(user_in_session.settings, key):
                        setattr(user_in_session.settings, key, value)
                
                await session.commit()
                await session.refresh(user_in_session.settings)
                logger.info(f"Modified settings for user ID {self.id}")
                return user_in_session.settings
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to modify settings for user ID {self.id}: {e}")
                return None

    async def modify_user_details(self, async_session: async_sessionmaker[AsyncSession], details_data: Dict[str, Any]) -> Optional[UserDetails]:
        from .user_details import UserDetails
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                if not user_in_session.details:
                    user_in_session.details = UserDetails(user_id=user_in_session.id)
                
                for key, value in details_data.items():
                    if hasattr(user_in_session.details, key):
                        setattr(user_in_session.details, key, value)
                
                await session.commit()
                await session.refresh(user_in_session.details)
                logger.info(f"Modified details for user ID {self.id}")
                return user_in_session.details
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to modify details for user ID {self.id}: {e}")
                return None

    async def add_user_action(self, async_session: async_sessionmaker[AsyncSession], action: UserAction) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                action_in_session = await session.merge(action)
                
                if action_in_session not in user_in_session.user_actions:
                    user_in_session.user_actions.append(action_in_session)
                    await session.commit()
                    logger.info(f"Added action '{action.name}' to user ID {self.id}")
                else:
                    logger.info(f"User ID {self.id} already has action '{action.name}'")

                return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to add action to user ID {self.id}: {e}")
                return None
    
    async def remove_device(self, async_session: async_sessionmaker[AsyncSession], action: UserAction) -> Optional['User']:
        pass
        