from aiokafka import AIOKafkaProducer
from app.config.kafka import get_kafka_producer
from app.models.requests.user import CreateUser
from app.models.schema.user import User
from app.models.responses import Token, UserPublic 
from app.services.auth_service import get_auth_service, AuthService
from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from typing import Any, Dict


router: APIRouter = APIRouter(
    prefix="/auth",
    tags=["controller responsible for user authentication"]
)


@router.post("/register", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def register_user(
    user_in: CreateUser,
    auth_service: AuthService = Depends(get_auth_service),
    producer: AIOKafkaProducer = Depends(get_kafka_producer)
):
 
    result = await auth_service.register_user(mail=user_in.mail, password=user_in.password)
    
    if not result:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User with this email already exists.",
        )
    
    new_user, activation_code = result
    
    await producer.send_and_wait(
        "user_activations", 
        {"user_id": new_user.id, "email": new_user.mail, "code": activation_code.code}
    )
    
    return new_user

@router.post("/activate/{activation_code}", response_model=UserPublic)
async def activate_account(
    activation_code: str,
    auth_service: AuthService = Depends(get_auth_service)
):

    user: User = await auth_service.activate_account(activation_code)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Invalid or expired activation code.",
        )
    return user

@router.post("/verify/{verification_code}", response_model=Token)
async def verify_account(
    verification_code: str,
    auth_service: AuthService = Depends(get_auth_service)
):

    result: Dict[str, Any] = await auth_service.verify_user_account(verification_code)
    if not result or result.get("status") == "failure":
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Invalid or expired activation code.",
        )
    return result

@router.post("/login", response_model=Token)
async def login_user(
    form_data: OAuth2PasswordRequestForm = Depends(),
    auth_service: AuthService = Depends(get_auth_service)
):
    result: Dict[str, Any] = await auth_service.login_user(mail=form_data.username, password=form_data.password)

    if not result or result.get("status") == "failure":
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=result.get("message", "Login failed"),
            headers={"WWW-Authenticate": "Bearer"},
        )
    return result


@router.post("/login/verify-2fa/{second_factor}", response_model=Token)
async def verify_2fa(
    second_factor_code: str,
    auth_service: AuthService = Depends(get_auth_service)
):
    result: Dict[str, Any] = await auth_service.verify_second_factor(
        second_factor_code=second_factor_code
    )

    if not result or result.get("status") != "success":
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired second-factor code.",
        )
    return result


