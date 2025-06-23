from pydantic import BaseModel, EmailStr
from typing import Optional

class CreateUser(BaseModel):
    mail: EmailStr
    password: str

    class Config:
        from_attributes = True
    
class ModifyUser(BaseModel):
    mail: EmailStr
    password: str

    class Config:
        from_attributes = True 
        
class CreateUserDetails(BaseModel):
    mail: EmailStr
    password: str

    class Config:
        from_attributes = True
  
        
class ModifyUserDetails(BaseModel):
    mail: EmailStr
    password: str

    class Config:
        from_attributes = True