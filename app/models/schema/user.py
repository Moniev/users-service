from __future__ import annotations
from app.config.database import Base
from app.config.redis import get_redis_client, Redis
from app.utils.encrypting_utils import encrypt_data, decrypt_data
from app.utils.redis_utils import get_cache, set_cache, revoke_cache
import asyncio
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError
from datetime import datetime
from fastapi import Depends
from loguru import logger
from redis.asyncio.client import Redis 
from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, Float, String, Table, Column, select, Select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from sqlalchemy.orm import Mapped, mapped_column, relationship, selectinload, aliased
from sqlalchemy.sql import func
from typing import Any, Dict, List, Optional, Tuple, TYPE_CHECKING, Union
import uuid as uuid_generator 


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
    black_listed: Mapped[bool] = mapped_column(Boolean, default=False)
    deleted: Mapped[bool] = mapped_column(Boolean, default=False)
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
        self.password: str = password
        if uuid is None:
            self.uuid = str(uuid_generator.uuid4()) 
        else:
            self.uuid = uuid 
        self.active: bool = active
        self.verified: bool = verified


    def __repr__(self):
        return f"<User id={self.id} mail='{self.mail}' active={self.active} verified={self.verified}>"


    def get_cache_keys(self) -> List[str]:
        keys: List[str] = []
        if self.id is not None:
            keys.append(f"user:id:{self.id}")
        if self.mail is not None:
            keys.append(f"user:mail:{self.mail}")
        if self.phone is not None:
            keys.append(f"user:phone:{self.phone}")
        return keys


    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "uuid": self.uuid,
            "mail": self.mail,
            "phone": self.phone,
            "password": self.password, 
            "active": self.active,
            "verified": self.verified,
            "black_listed": self.black_listed,
            "deleted": self.deleted,
            "details": self.details.to_dict() if self.activation_code else None,
            "devices": [device.to_dict() for device in self.devices],
            "user_roles": [role.to_dict() for role in self.user_roles],
            "user_actions": [action.to_dict() for action in self.user_actions],
            "settings": self.settings.to_dict() if self.activation_code else None,
            "reset_code": self.reset_code.to_dict() if self.reset_code else None,
            "activation_code": self.activation_code.to_dict() if self.activation_code else None,
            "second_factor_code": self.second_factor_code.to_dict() if self.second_factor_code else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> Optional["User"]:
        if 'password' not in data:
            logger.error("Attempted to restore user from cache, but password was missing.")
            return None
       
        user: User = cls(
            mail=data['mail'],
            phone=data['phone'],
            password=data['password'],
            uuid=data.get('uuid'),
            active=data.get('active', False),
            verified=data.get('verified', False)
        )
        user.id = data['id']
        user.black_listed = data.get('black_listed', False)
        user.deleted = data.get('deleted', False)
        user.created_at = datetime.fromisoformat(data['created_at']) if data.get('created_at') else None
        user.updated_at = datetime.fromisoformat(data['updated_at']) if data.get('updated_at') else None
        if data.get('details'):
            from .user_details import UserDetails 
            user.details = UserDetails.from_dict(data['details']) 
        if data.get('devices'):
            from .user_device import UserDevice
            user.devices = [UserDevice.from_dict(d) for d in data['devices']]
        return user


    @classmethod
    async def create(cls, mail: str, password: str, phone: Optional[str] = None) -> User:
        hashed_password = await asyncio.to_thread(password_hasher.hash, password)
        return cls(mail=mail, password=hashed_password, phone=phone)


    @classmethod
    async def get_user_by_id(cls, async_session: async_sessionmaker[AsyncSession], id: int, redis_client: Redis) -> Optional[User]:
        cache_key: str = f"user:id:{id}"
        user_data: Dict[str, Any] = await get_cache(cache_key, redis_client) 
        if user_data: 
            logger.info(f"Cache hit for user ID: {id}.")
            user: User = User.from_dict(user_data) 
            return user
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .where(User.id == id)
                .options(
                    selectinload(User.details),
                    selectinload(User.devices),
                    selectinload(User.user_roles),
                    selectinload(User.user_actions),
                    selectinload(User.settings),
                    selectinload(User.reset_code),
                    selectinload(User.activation_code),
                    selectinload(User.second_factor_code),
                    selectinload(User.verification_code)
                )
            )
            result = await session.execute(statement)
            user: User =  result.scalars().first()
            if user:
                await set_cache(user.to_dict(), user.get_cache_keys(), redis_client)
                logger.info(f"User ID: {id} fetched from DB and cached.")
            
            return user
    
    
    @classmethod
    async def get_user_by_mail(cls, async_session: async_sessionmaker[AsyncSession], mail: str, redis_client: Redis) -> Optional[User]:
        cache_key: str = f"user:mail:{mail}"
        user_dict: Dict[str, Any] = await get_cache(cache_key, redis_client)
        if user_dict:
            user: User = User.from_dict(user_dict)
            return user
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .where(User.mail == mail)
                .options(
                    selectinload(User.details),
                    selectinload(User.devices),
                    selectinload(User.user_roles),
                    selectinload(User.user_actions),
                    selectinload(User.settings),
                    selectinload(User.reset_code),
                    selectinload(User.activation_code),
                    selectinload(User.second_factor_code),
                    selectinload(User.verification_code)
                )
            )
            result = await session.execute(statement)
            user: User =  result.scalars().first()
            if user:
                await set_cache(user, user.get_cache_keys(), redis_client)
                
            return user
    
    
    @classmethod
    async def get_user_by_activation_code(cls, async_session: async_sessionmaker[AsyncSession], code: str, redis_client: Redis) -> Optional[User]:
        from app.models.schema import ActivationCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.activation_code) 
                .where(ActivationCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            return result.scalars().first()
    
    
    @classmethod
    async def get_user_by_verification_code(cls, async_session: async_sessionmaker[AsyncSession], code: str, redis_client: Redis) -> Optional[User]:
        from app.models.schema import VerificationCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.verification_code) 
                .where(VerificationCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            return result.scalars().first()
    
    
    @classmethod
    async def get_user_by_second_factor_code(cls, async_session: async_sessionmaker[AsyncSession], code: str, redis_client: Redis) -> Optional[User]:
        from app.models.schema import SecondFactorCode
        
        async with async_session() as session:
            statement: Select = select(User).options(selectinload(User.settings)).join(User.second_factor_code).where(SecondFactorCode.code == code)
            result = await session.execute(statement)
            user: User = result.scalars().first()
            
            return user
        
        
    @classmethod
    async def get_user_by_reset_code(cls, async_session: async_sessionmaker[AsyncSession], code: str, redis_client: Redis) -> Optional[User]:
        from app.models.schema import ResetCode
        
        async with async_session() as session:
            statement: Select = (
                select(User)
                .join(User.reset_code) 
                .where(ResetCode.code == code) 
            )
            
            result = await session.execute(statement)
            
            return result.scalars().first()
    
    
    @classmethod
    async def verify_password(cls, async_session: async_sessionmaker[AsyncSession], mail: str, password: str, redis_client: Redis) -> Optional[User]:
        async with async_session() as session:
            statement: Select = select(User).options(
            selectinload(User.details),
                selectinload(User.devices),
                selectinload(User.user_roles),
                selectinload(User.user_actions),
                selectinload(User.settings),
                selectinload(User.reset_code),
                selectinload(User.activation_code),
                selectinload(User.second_factor_code),
                selectinload(User.verification_code)
            ).where(User.mail == mail)
            result = await session.execute(statement)
            user: User = result.scalars().first()

            if not user:
                logger.info(f"User with mail: {mail} not found.")
                return None

            try:
                await asyncio.to_thread(password_hasher.verify, user.password, password)
                
                logger.info(f"Successfully verified user with ID: {user.id}")
                return user

            except VerifyMismatchError:
                logger.info(f"Provided wrong password for user with ID: {user.id}")
                return None
            except Exception as e:
                    await session.rollback()
                    logger.error(f"Error during password verification for user ID {user.id}: {e}")
                    return None
    
    
    @classmethod
    async def register(cls, async_session: async_sessionmaker[AsyncSession], mail: str, password: str, redis_client: Redis) -> Optional[Tuple[User, ActivationCode]]:
        from .activation_code import ActivationCode
        async with async_session() as session:
            existing_user = (await session.execute(select(User).where(User.mail == mail))).scalars().first()
            if existing_user:
                logger.warning(f"Mail '{mail}' already exists in database. Registration aborted.") 
                return None
            try:
                user = await User.create(mail=mail, password=password, phone=None)
                session.add(user)
                await session.flush() 
                activation_code = ActivationCode.create_activation_code(user_id=user.id) 

                if not isinstance(activation_code, ActivationCode):
                    logger.error(f"ActivationCode.create_activation_code() returned an invalid type ({type(activation_code)}) or None. Expected ActivationCode instance. Registration aborted for {mail}. Please check the implementation of ActivationCode.py.") # Enhanced log
                    await session.rollback()
                    return None

                user.activation_code = activation_code 
                
                logger.debug(f"Attempting to add user (ID: {getattr(user, 'id', 'N/A')}, Mail: {user.mail}) "
                             f"with ActivationCode (Code: {getattr(user.activation_code, 'code', 'N/A')}, Type: {type(user.activation_code)}) to session.")
                
                session.add(activation_code) 
                
                await session.commit() 
                await session.refresh(user, attribute_names=[
                    "details", "devices", "user_roles", "user_actions", "settings",
                    "reset_code", "activation_code", "second_factor_code", "verification_code"
                ])
                
                await session.refresh(user) 
                await session.refresh(activation_code) 
                logger.info(f"Successfully created user with ID: {user.id}")
                
                await set_cache(user, user.get_cache_keys(), redis_client)
                return user, activation_code
            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Integrity error during registration for email {mail}: {e}")
                return None
            except Exception as e:
                await session.rollback()
                logger.error(f"Unexpected error during registration for email {mail}: {e}")
                return None


    async def activate_account(self, async_session: async_sessionmaker[AsyncSession], activation_code: Union[str, ActivationCode], redis_client: Redis) -> Optional['User']:
        from app.models.schema import ActivationCode
        
        async with async_session() as session:
            user_in_session: User = await session.merge(self)

            if user_in_session.active:
                logger.info(f"User {user_in_session.id} is already activated.")
                return user_in_session
            
            await session.refresh(user_in_session, attribute_names=["activation_code"])
            
            code_str = activation_code.code if isinstance(activation_code, ActivationCode) else activation_code

            if not user_in_session.activation_code or user_in_session.activation_code.code != code_str:
                logger.error(f"Failed to activate user with ID: {user_in_session.id}. Invalid or missing activation code.")
                return None

            try:
                user_in_session.active = True
                await session.delete(user_in_session.activation_code) 

                await session.commit()
                await session.refresh(user_in_session) 

                logger.info(f"Successfully activated user with ID: {user_in_session.id}")
                await set_cache(user_in_session, user_in_session.get_cache_keys(), redis_client)
                return user_in_session

            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Integrity error for user with ID: {user_in_session.id}. {e}")
                return None
            except Exception as e:
                await session.rollback()
                logger.error(f"Unknown error for user with ID: {user_in_session.id}: {e}")
                return None
        

    async def verify_account(self, async_session: async_sessionmaker[AsyncSession], verification_code: Union[str, VerificationCode], redis_client: Redis) -> Optional['User']:
        from app.models.schema import VerificationCode
        
        async with async_session() as session:
            user_in_session: User = await session.merge(self)

            if user_in_session.verified:
                logger.info(f"User {user_in_session.id} is already verified.")
                return user_in_session
            
            await session.refresh(user_in_session, attribute_names=["verification_code"])

            code_str = verification_code.code if isinstance(verification_code, VerificationCode) else verification_code

            if not user_in_session.verification_code or user_in_session.verification_code.code != code_str:
                logger.error(f"Failed to verify user with ID: {user_in_session.id}. Invalid or missing verification code.")
                return None

            try:
                user_in_session.verified = True
                await session.delete(user_in_session.verification_code)

                await session.commit()
                await session.refresh(user_in_session)

                logger.info(f"Successfully verified user with ID: {user_in_session.id}")
                await set_cache(user_in_session, user_in_session.get_cache_keys(), redis_client)
                return user_in_session

            except IntegrityError as e:
                await session.rollback()
                logger.error(f"Integrity error for user with ID: {user_in_session.id}. {e}")
                return None
            except Exception as e:
                await session.rollback()
                logger.error(f"Unknown error for user with ID: {user_in_session.id}: {e}")
                return None
             
                
    async def create_verification_code(self, async_session: async_sessionmaker[AsyncSession], redis_client: Redis) -> Optional[VerificationCode]:
        from .verification_code import VerificationCode
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                if user_in_session.verified:
                    logger.info(f"User with ID {user_in_session.id} is already verified. No new code created.")
                    return user_in_session.verification_code 
                    
                await session.refresh(user_in_session, attribute_names=["verification_code"])
                if user_in_session.verification_code:
                    await session.delete(user_in_session.verification_code)
                
                new_code: VerificationCode = VerificationCode.create_verification_code()
                user_in_session.verification_code = new_code
                await session.commit()
                await session.refresh(new_code)
                await set_cache(user_in_session, user_in_session.get_cache_keys(), redis_client)
                logger.info(f"Created new verification code for user ID {user_in_session.id}")
                return new_code
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to create verification code for user ID {self.id}: {e}")
                return None


    async def generate_second_factor_code(self, async_session: async_sessionmaker[AsyncSession]) -> Optional['SecondFactorCode']:
        from .second_factor_code import SecondFactorCode
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                
                await session.refresh(user_in_session, attribute_names=["second_factor_code"])
                if user_in_session.second_factor_code:
                    await session.delete(user_in_session.second_factor_code)
                
                code_string: str = SecondFactorCode.create_second_factor_code() 
                
                new_2fa_code = SecondFactorCode(code=code_string, user_id=user_in_session.id)

                user_in_session.second_factor_code = new_2fa_code 
                
                session.add(new_2fa_code) 
                
                await session.commit()
                await session.refresh(new_2fa_code)
                logger.success(f"Generated 2FA code {new_2fa_code.code[:4]}... for user ID: {user_in_session.id}")
                await set_cache(user_in_session, user_in_session.get_cache_keys())
                return new_2fa_code
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to generate 2FA code for user ID {self.id}: {e}")
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
                
                await set_cache(user_in_session, user_in_session.get_cache_keys())
                return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to add role to user ID {self.id}: {e}")
                return None


    async def revoke_user_role(self, async_session: async_sessionmaker[AsyncSession], role: UserRole) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                await session.refresh(user_in_session, attribute_names=["user_roles"])
                
                role_to_revoke = next((r for r in user_in_session.user_roles if r.id == role.id), None)

                if role_to_revoke:
                    user_in_session.user_roles.remove(role_to_revoke)
                    await session.commit()
                    logger.info(f"Revoked role '{role.name}' from user ID {self.id}")
                else:
                    logger.warning(f"Role '{role.name}' not found for user ID {self.id}. Cannot revoke.")
                
                await set_cache(user_in_session, user_in_session.get_cache_keys())
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
                await session.refresh(user_in_session, attribute_names=["settings"])

                if not user_in_session.settings:
                    user_in_session.settings = UserSettings(user_id=user_in_session.id)
                
                for key, value in settings_data.items():
                    if hasattr(user_in_session.settings, key):
                        setattr(user_in_session.settings, key, value)
                
                await session.commit()
                await session.refresh(user_in_session.settings)
                logger.info(f"Modified settings for user ID {self.id}")
                await set_cache(user_in_session, user_in_session.get_cache_keys())
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
                await session.refresh(user_in_session, attribute_names=["details"])

                if not user_in_session.details:
                    user_in_session.details = UserDetails(user_id=user_in_session.id)
                
                for key, value in details_data.items():
                    if hasattr(user_in_session.details, key):
                        setattr(user_in_session.details, key, value)
                
                await session.commit()
                await session.refresh(user_in_session.details)
                logger.info(f"Modified details for user ID {self.id}")
                await set_cache(user_in_session, user_in_session.get_cache_keys())
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
                
                await session.refresh(user_in_session, attribute_names=["user_actions"])

                if action_in_session not in user_in_session.user_actions:
                    user_in_session.user_actions.append(action_in_session)
                    await session.commit()
                    logger.info(f"Added action '{action.name}' to user ID {self.id}")
                else:
                    logger.info(f"User ID {self.id} already has action '{action.name}'")

                await set_cache(user_in_session, user_in_session.get_cache_keys())
                return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to add action to user ID {self.id}: {e}")
                return None
    
    
    async def remove_device(self, async_session: async_sessionmaker[AsyncSession], device_id: int, redis_client: Redis) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session: User = await session.merge(self)
                await session.refresh(user_in_session, attribute_names=["devices"])
                
                device_to_remove: Optional[UserDevice] = next((d for d in user_in_session.devices if d.id == device_id), None)
                
                if device_to_remove:
                    device_to_remove.deleted = True 
                    await session.add(device_to_remove) 
                    await session.commit()
                    await session.refresh(user_in_session) 
                    logger.success(f"Successfully marked device with ID: {device_id} as deleted for user ID {user_in_session.id}")
                    await set_cache(user_in_session, user_in_session.get_cache_keys(), redis_client)
                    return user_in_session
                else:
                    logger.warning(f"Device with ID: {device_id} not found for user ID {user_in_session.id}.")
                    return user_in_session
            except Exception as e:
                await session.rollback()
                logger.error(f"Failed to remove device for user ID {self.id}: {e}")
                return None
    
    
    async def remove_second_factor(self, async_session: async_sessionmaker[AsyncSession], redis_client: Redis) -> Optional['User']:
        async with async_session() as session:
            try:
                user_in_session = await session.merge(self)
                await session.refresh(user_in_session, attribute_names=["second_factor_code"])

                if user_in_session.second_factor_code:
                    await session.delete(user_in_session.second_factor_code)
                    await session.commit()
                    await session.refresh(user_in_session) 
                    logger.success(f"2FA code successfully removed for user ID: {user_in_session.id}.")
                    
                    await set_cache(user_in_session, user_in_session.get_cache_keys(), redis_client)
                    return user_in_session
                else:
                    logger.warning(f"No 2FA code found to remove for user ID {user_in_session.id}.")
                    return user_in_session 
            except Exception as e:
                await session.rollback()
                logger.error(f"An error occurred during 2FA code removal for user ID {self.id}: {e}")
                return None
