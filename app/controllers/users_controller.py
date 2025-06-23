from fastapi import APIRouter

router: APIRouter = APIRouter(
    prefix="/users",
    tags=["Users management"]
)

@router.get("/get-users/public", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def get_users_public():
    pass

@router.get("/get-users/private", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def get_users_private():
    pass

@router.get("/get-user/public", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def get_user_public():
    pass

@router.get("/get-user/private", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def get_user_private():
    pass

@router.post("/user-details/add", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def add_details():
    pass

@router.put("/user-details/reset", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def reset_details():
    pass

@router.patch("/user-details/update", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def update_details():
    pass

@router.put("/settings/reset", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def reset_settings():
    pass

@router.patch("/settings/update", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def update_settings():
    pass