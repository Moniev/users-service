import os
import jwt
from datetime import datetime, timezone
from typing import Any, Dict, Optional, List
from fastapi import FastAPI, Request, Response, HTTPException, status
from starlette.middleware.base import BaseHTTPMiddleware, RequestResponseEndpoint
from starlette.responses import JSONResponse
from app.config.settings import settings
from loguru import logger

class JWTAuthMiddleware(BaseHTTPMiddleware):
    def __init__(self, app: FastAPI):
        super().__init__(app)
        self.jwt_public_key: Optional[str] = None
        self.jwt_algorithm: str = "EdDSA"

        self.load_jwt_public_key()


    def load_jwt_public_key(self):
        public_key_path = settings.JWT_PUBLIC_KEY_PATH

        if not public_key_path:
            logger.error("JWT public key path is not defined in settings (settings.JWT_PUBLIC_KEY_PATH).")
            if os.environ.get("TEST_MODE") != "True":
                raise ValueError("JWT public key path is required in this application mode.")
            else:
                logger.warning("No JWT public key path in test mode. Token generation might not be possible unless keys are generated in AuthService.")
                return
        
        try:
            with open(public_key_path, 'r') as f:
                self.jwt_public_key = f.read()
            logger.info("Successfully loaded JWT public key from file.")
        except FileNotFoundError as e:
            logger.error(f"Error loading JWT public key: File not found. Make sure the secret is mounted properly: {e}")
            raise FileNotFoundError(f"JWT public key file not found: {e}")
        except Exception as e:
            logger.error(f"Unexpected error while loading JWT public key: {e}")
            raise Exception(f"Failed to configure JWT public key: {e}")


    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        if request.url.path.startswith(("/docs", "/redoc", "/openapi.json")):
            return await call_next(request)

        if os.environ.get("TEST_MODE") == "True" and not self.jwt_public_key:
            logger.warning("JWT middleware is operating in test mode, but the public key has not been loaded. Attempting to use key from AuthService...")
            try:
                from app.services.auth_service import AuthService
                logger.warning("Cannot dynamically retrieve public key from AuthService in middleware. Please ensure JWT_PUBLIC_KEY_PATH is set in test mode or mock jwt.decode.")
            except ImportError:
                logger.warning("AuthService cannot be imported into middleware. JWT verification in test mode will be impossible without a defined public key.")

        auth_header = request.headers.get("Authorization")
        if not auth_header:
            logger.warning("Missing Authorization header.")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Authentication token missing",
                headers={"WWW-Authenticate": "Bearer"},
            )

        try:
            scheme, token = auth_header.split()
            if scheme.lower() != "bearer":
                logger.warning("Invalid authorization scheme.")
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Authentication scheme must be 'Bearer'",
                    headers={"WWW-Authenticate": "Bearer"},
                )
            
            if not self.jwt_public_key:
                logger.error("JWT public key not loaded. Cannot verify token.")
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail="JWT server configuration error",
                )


            payload: Any = jwt.decode(token, self.jwt_public_key, algorithms=[self.jwt_algorithm])

            if payload.get("exp") is None:
                logger.warning("Missing expiration time (exp) in JWT payload.")
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Incomplete JWT token (missing 'exp')",
                    headers={"WWW-Authenticate": "Bearer"},
                )
            
            expire_time = datetime.fromtimestamp(payload["exp"], tz=timezone.utc)
            if expire_time < datetime.now(timezone.utc):
                logger.warning(f"JWT token expired. Expiration time: {expire_time}, Current time: {datetime.now(timezone.utc)}")
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="JWT token expired",
                    headers={"WWW-Authenticate": "Bearer"},
                )


            request.state.user = {
                "user_id": payload.get("sub"),
                "mail": payload.get("mail"),
                "user_roles": payload.get("user_roles", [])
            }
            
            logger.debug(f"Authenticated user: {request.state.user['mail']} (ID: {request.state.user['user_id']})")

        except jwt.ExpiredSignatureError:
            logger.warning("Expired JWT token.")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="JWT token expired",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except jwt.InvalidTokenError as e:
            logger.warning(f"Invalid JWT token: {e}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Invalid authentication token: {e}",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except ValueError as e:
            logger.warning(f"Error parsing Authorization header: {e}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid Authorization header format",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except Exception as e:
            logger.error(f"Unexpected error in JWT middleware: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Internal authentication server error",
            )

        response = await call_next(request)
        return response
