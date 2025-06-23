from aiokafka import AIOKafkaProducer
from app.config.kafka import get_kafka_producer
from app.config.database import get_session_factory
from app.models.requests.user import CreateUser
from app.models.responses.user import UserPublic
from app.models.schema.user import User
from app.models.schema.activation_code import ActivationCode
from datetime import timedelta
from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from typing import Tuple, Optional


router: APIRouter = APIRouter(
    prefix="/auth",
    tags=["controller responsible for user authentication"]
)


@router.post("/register", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def register_user(
    user_in: CreateUser,
    session_factory: async_sessionmaker[AsyncSession] = Depends(get_session_factory),
    producer: AIOKafkaProducer = Depends(get_kafka_producer)
):
   
    result: Optional[Tuple[User, ActivationCode]] = await User.register(
        async_session=session_factory, 
        mail=user_in.mail, 
        password=user_in.password
    )
    
    if not result:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User with this email already exists.",
        )
    
    new_user, activation_code = result

    
    return new_user

@router.post("/activate", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def activate():
    pass

@router.post("/login", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def activate():
    pass

@router.post("/verify", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def activate():
    pass

@router.patch("/reset-password", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def activate():
    pass

@router.patch("/change-password", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def activate():
    pass