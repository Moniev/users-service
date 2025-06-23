from pydantic import BaseModel, EmailStr
from typing import Optional

class UserPublic(BaseModel):
    id: int
    uuid: Optional[str]
    mail: EmailStr
    active: bool
    verified: bool

    class Config:
        from_attributes = True
        
class UserPrivate(BaseModel):
    id: int
    uuid: Optional[str]
    mail: EmailStr
    active: bool
    verified: bool

    class Config:
        from_attributes = True
