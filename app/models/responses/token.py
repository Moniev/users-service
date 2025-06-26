from pydantic import BaseModel
from typing import Optional


class UserSettings(BaseModel):
    theme: Optional[str]
    notifications_enabled: Optional[bool]
    notifications_login_enabled: Optional[bool]
    notifications_personal_enabled: Optional[bool]
    notifications_tasks_enabled: Optional[bool]
    second_factor_enabled: Optional[bool]
    
    class Config:
        from_attributes = True


class Token(BaseModel):
    status: Optional[str]
    token: Optional[str]
    user_settings: Optional[UserSettings]

    class Config:
        from_attributes = True

