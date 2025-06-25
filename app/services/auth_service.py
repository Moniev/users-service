import os
from app.config.settings import settings
from app.models.schema import User, ActivationCode, ResetCode, VerificationCode, SecondFactorCode, UserAction, UserDetails, UserDevice, UserRole, UserSettings
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import ed25519
from cryptography.hazmat.primitives.serialization import Encoding, PrivateFormat, PublicFormat, NoEncryption
from datetime import datetime, timedelta, timezone 
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker
from redis.asyncio.client import Redis
from aiokafka import AIOKafkaProducer
import jwt
from loguru import logger
from typing import Any, Dict, Optional, Tuple


class AuthService:
    def __init__(self, session_factory: async_sessionmaker[AsyncSession], kafka_producer: AIOKafkaProducer = None, redis_client: Redis = None, ):
        self.async_session: async_sessionmaker[AsyncSession] = session_factory
        self.kafka_producer: AIOKafkaProducer = kafka_producer
        self.redis_client: Redis = redis_client
        self.private_key_path: str = settings.JWT_PRIVATE_KEY_PATH
        self.public_key_path: str = settings.JWT_PUBLIC_KEY_PATH
        self.jwt_private_key: Optional[str] = None
        self.jwt_public_key: Optional[str] = None

        if os.environ.get("TEST_MODE") != "True":
            if not self.private_key_path or not self.public_key_path:
                logger.error("JWT keys paths are not defined")
                raise ValueError("JWT keys paths are required in this application mode")
            
            try:
                with open(self.private_key_path, 'r') as f:
                    self.jwt_private_key = f.read()
                with open(self.public_key_path, 'r') as f:
                    self.jwt_public_key = f.read()
                logger.info("Successfully loaded JWT keys")
            except FileNotFoundError as e:
                logger.error(f"Failed to load JWT keys: Files not found. Make sure secret is mounted properly {e}")
                raise FileNotFoundError(f"JWT keys not found: {e}")
            except Exception as e:
                logger.error(f"Unexpected error while loading JWT keys: {e}")
                raise Exception(f"Failed to config JWT keys: {e}")

        else:
            logger.info("Application is running in testing mode, generating new JWT keys")
            try:
                private_key_obj = ed25519.Ed25519PrivateKey.generate()
                public_key_obj = private_key_obj.public_key()

                self.jwt_private_key = private_key_obj.private_bytes(
                    encoding=Encoding.PEM,
                    format=PrivateFormat.PKCS8,
                    encryption_algorithm=NoEncryption()
                ).decode('utf-8')
                
                self.jwt_public_key = public_key_obj.public_bytes(
                    encoding=Encoding.PEM,
                    format=PublicFormat.SubjectPublicKeyInfo
                ).decode('utf-8')
                logger.success("Successfully generated JWT keys is testing mode")
            except Exception as e:
                logger.error(f"Error while generating JWT keys: {e}")
                raise Exception(f"Failed to generate JWT keys in test mode: {e}")

        self.jwt_algorithm = "EdDSA"


    async def generate_jwt_token(self, user_id: int, user_mail: str, expires_delta: Optional[timedelta] = None) -> str:
        if expires_delta:
            expire = datetime.now(timezone.utc) + expires_delta 
        else:
            expire = datetime.now(timezone.utc) + timedelta(hours=24) 

        user_roles: UserRole =  await UserRole.get_user_role_by_user_id(self.async_session, user_id)
        to_encode: Dict[str, Any] = {"sub": str(user_id), "mail": user_mail, "exp": expire, "user_roles": user_roles}
        
        encoded_jwt = jwt.encode(to_encode, self.jwt_private_key, algorithm=self.jwt_algorithm)
        logger.info(f"Generated JWT token for use with ID: {user_id}, which expires at: {expire}")
        return encoded_jwt


    async def register_user(self, mail: str, password: str) -> Optional[Tuple[User, ActivationCode]]:
        logger.info(f"Attempting to register new user with email: {mail}")
        exisiting_user: User = await User.get_user_by_mail(self.async_session, mail)
        if exisiting_user:
            logger.info("user with provided email already exists")
            return None
        
        registration_result = await User.register(self.async_session, mail, password)
        if registration_result:
            user, activation_code = registration_result
            logger.success(f"User {mail} registered successfully.")
            return user, activation_code
        else:
            logger.warning(f"Registration failed for email: {mail}. It might already exist.")
            return None


    async def login_user(self, mail: str, password: str) -> Optional[Dict[str, Any]]:
        logger.info(f"Login attempt for user: {mail}")
        user = await User.verify_password(self.async_session, mail, password)
        
        if not user:
            logger.warning(f"Login failed for {mail}: Invalid credentials.")
            return {"status": "failure", "message": "Invalid credentials."}

        if not user.active:
            logger.warning(f"Login failed for {mail}: Account is not active.")
            return {"status": "failure", "message": "Account is not active."}

        if user.settings and user.settings.second_factor_enabled:
            logger.info(f"Two-factor authentication is enabled for user: {mail}. Generating code.")
            second_factor_code: SecondFactorCode = await user.generate_second_factor_code(self.async_session)
            
            try:
                logger.success(f"Generated 2FA code for user: {mail}")
                return {"status": "mfa_required", "user_id": user.id}
            except Exception as e:
                logger.error(f"Failed to generate 2FA code for user {mail}: {e}")
                return {"status": "failure", "message": "Failed to process 2FA."}

        logger.success(f"Login successful for user: {mail} (2FA not enabled).")
        jwt_token = self.generate_jwt_token(user.id, user.mail)
        return {"status": "success", "user_id": user.id, "token": jwt_token, "user_settings": user.settings}


    async def verify_second_factor(self, second_factor_code: str) -> Optional[User]:
        logger.info(f"Attempting to verify 2FA code: {second_factor_code[:4]}...")
        user: User = await User.get_user_by_second_factor_code(self.async_session, second_factor_code)

        if not user:
            logger.warning("2FA verification failed: Code is invalid or expired.")
            return None

        user = await user.remove_second_factor(self.async_session)
        jwt_token = self.generate_jwt_token(user.id, user.mail)
        
        return {"status": "success", "user_id": user.id, "token": jwt_token, "user_settings": user.settings}


    async def activate_account(self, activation_code_str: str) -> Optional[User]:
        logger.info(f"Attempting to activate account with code: {activation_code_str[:8]}...")
        user: User = await User.get_user_by_activation_code(self.async_session, activation_code_str)
        if not user:
            logger.error("Activation failed: Code not found or invalid.")
            return None

        user = await user.activate_account(self.async_session, activation_code_str) 
        return user


    async def create_verification_code(self, user_id: int) -> Optional[VerificationCode]:
        logger.info(f"Requesting verification code for user ID: {user_id}")
        user: User = await User.get_user_by_id(self.async_session, user_id)
        if not user:
            logger.error(f"Cannot create verification code: User with ID {user_id} not found.")
            return None
        
        verification_code = await user.create_verification_code(self.async_session)
        return verification_code


    async def verify_user_account(self, verification_code_str: str) -> Optional[User]:
        logger.info(f"Attempting to verify account with code: {verification_code_str[:8]}...")
        user = await User.get_user_by_verification_code(self.async_session, verification_code_str)
        if not user:
            logger.error("Verification failed: Code not found or invalid.")
            return None

        verification_code_obj = VerificationCode(code=verification_code_str, user_id=user.id)
        user = await user.verify_account(self.async_session, verification_code_obj)
        return user
        
        
    async def get_user_by_id(self, user_id: int) -> Optional[User]:
        return await User.get_user_by_id(self.async_session, user_id)


    async def remove_device(self, user_id: int, device_id: str) -> Optional[User]:
        logger.info(f"Attempting to remove device '{device_id}' from user ID {user_id}")
        user = await User.get_user_by_id(self.async_session, user_id)
        if not user:
            logger.error(f"Cannot remove device: User with ID {user_id} not found.")
            return None

        user = await user.remove_device(self.async_session, device_id=device_id)
        return user
