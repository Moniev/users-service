from app.models.schema.activation_code import ActivationCode
from app.models.schema.second_factor_code import SecondFactorCode
from app.models.schema.reset_code import ResetCode
from app.models.schema.user import User
from app.models.schema.user_device import UserDevice
from app.models.schema.user_role import UserRole
from app.models.schema.user_settings import UserSettings
from app.models.schema.verification_code import VerificationCode
from app.config.database import Base
from app.config.logger import setup_logger
import json, pytest, pytest_asyncio, yaml
from loguru import logger
from pathlib import Path
from pydantic import BaseModel 
from sqlalchemy import create_engine, select, func
from sqlalchemy.orm import Session, selectinload, joinedload
from typing import Any, Generator, Callable, Tuple, Dict, List
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession


async def verify_row_count(async_session: async_sessionmaker[AsyncSession], params: list):
    async with async_session() as session: 
        for check in params:
            model_class = SCHEMA_MAP.get(check["model"])
            statement = select(func.count()).select_from(model_class)
            result = await session.execute(statement)
            actual_count = result.scalar_one()
            expected_count = check["expected_count"]
            assert actual_count == expected_count
    logger.success("Row counts verified successfully.")


async def verify_user_details(
    async_session: async_sessionmaker[AsyncSession],
    user_id: int, 
    expected_data: dict):

    async with async_session() as session:
        user = await session.get(User, user_id)

    assert user is not None
    for key, expected_value in expected_data.items():
        actual_value = getattr(user, key)
        assert actual_value == expected_value, f"User {user_id} field '{key}' has value '{actual_value}', expected '{expected_value}'"
    logger.success(f"Details for user {user_id} are correct.")


async def verify_user_and_relations(
    async_session: async_sessionmaker[AsyncSession], 
    user_id: int, 
    expected_email: str, 
    expected_device_count: int, 
    expected_theme: str):
    async with async_session() as session:
        statement = (
            select(User)
            .where(User.id == user_id)
            .options(
                selectinload(User.devices),    
                joinedload(User.settings)        
            )
        )
        
        result = await session.execute(statement)
        user = result.scalar_one_or_none()

    assert user is not None, f"User with ID {user_id} not found."
    assert user.mail == expected_email
    assert user.settings is not None, f"UserSettings for user {user_id} not found."
    assert user.settings.theme == expected_theme
    assert len(user.devices) == expected_device_count
    
    logger.success(f"All relations for user {user_id} verified successfully.")


FUNCTION_MAP: dict[str,  Callable[..., Any]] = {
    "user_get_user_by_id": User.get_user_by_id,
    "user_get_user_by_mail": User.get_user_by_mail,
    "user_verify_password": User.verify_password,
    "user_register": User.register,
    "setup_logger": setup_logger,
    "verify_row_count": verify_row_count,
    "verify_user_details": verify_user_details,
    "verify_user_and_relations": verify_user_and_relations,
    "user_get_user_by_id": User.get_user_by_id,
    "user_get_user_by_mail": User.get_user_by_mail,
    "user_get_user_by_activation_code": User.get_user_by_activation_code
}


SCHEMA_MAP: dict[str, Base] = {
    "ActivationCode": ActivationCode,
    "ResetCode": ResetCode,
    "SecondFactorCode": SecondFactorCode,
    "User": User,
    "UserDevice": UserDevice,
    "UserRole": UserRole,
    "UserSettings": UserSettings,
    "VerificationCode": VerificationCode,
}

BASE_PATH: Path = Path(__file__).parent

RESPONSE_MAP: dict[str, BaseModel] = {}

RESOURCE_PATH: Path = BASE_PATH / "resources"
SETTINGS_PATH: Path = BASE_PATH / "settings" / "settings.yaml"
LOGS_PATH: Path = BASE_PATH / "logs" 

INTEGRATION_TEST_PATH: Path = RESOURCE_PATH / "integration"
UNIT_TEST_PATH: Path = RESOURCE_PATH / "unit"
E2E_TEST_PATH: Path = RESOURCE_PATH / "e2e"

TEST_FILES: list[Path] = list(INTEGRATION_TEST_PATH.glob("*.json"))
TEST_FILES.extend(list(UNIT_TEST_PATH.glob("*.json")))
TEST_FILES.extend(list(E2E_TEST_PATH.glob("*.json")))


setup_logger(LOGS_PATH)


@pytest.fixture(scope="session")  
def app_settings() -> Any:
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    
    if not SETTINGS_PATH.exists():
        raise FileNotFoundError(f"failed to found settings.yaml in path {SETTINGS_PATH}")
    
    with open(SETTINGS_PATH, 'r') as f:
        settings: Any = yaml.safe_load(f)
    
    return settings


@pytest.fixture(scope="function")
def create_fixed_db_session(type: str) -> Generator[Session, Any, Any]:
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    
    session: Session = Session(engine)

    def load_data(json_file_name: str):     
        resource_path: Path = RESOURCE_PATH / type / json_file_name
        with open(resource_path, 'r') as f:
            data = json.load(f)

        for item in data:
            model_class = SCHEMA_MAP.get(item["model"])
            if not model_class:
                raise ValueError(f"unknown model: '{item['model']}' in file.")
            
            instance: Any = model_class(**item["data"])
            session.add(instance)
        
        session.commit()
        return session

    yield load_data

    session.close()
    Base.metadata.drop_all(engine)


@pytest.fixture(scope="function")
def create_empty_db_session() -> Generator[Session, Any, Any]:
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    
    with Session(engine) as session:
        yield session
    
    Base.metadata.drop_all(engine)
    
    
@pytest.fixture(scope="function")
def create_data_driven_db_session() -> Generator[Callable[[str], Tuple[Session, List[Dict]]], None, None]:
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    
    session: Session = Session(engine)

    def load_data_and_test_cases(type: str, json_file_name: str) -> Tuple[Session, List[Dict]]:
        resource_path: Path = RESOURCE_PATH / type / json_file_name
        
        if not resource_path.exists():
            raise FileNotFoundError(f"Could not find fixture file: {resource_path}")
            
        with open(resource_path, 'r') as f:
            data: Dict = json.load(f)

        setup_data: List[Dict] = data.get("setup_data", [])
        for item in setup_data:
            model_class = SCHEMA_MAP.get(item["model"])
            if not model_class:
                raise ValueError(f"Unknown model: '{item['model']}' in JSON file.")
            
            instance_data: Dict[str] = item["data"].copy()
            instance_id: int = instance_data.pop('id', None)
            instance: Base = model_class(**instance_data)
            if instance_id is not None:
                instance.id = instance_id
            session.add(instance)
        
        session.commit()
        session.close()
        
        test_cases = data.get("test_cases", [])
        
        return session, test_cases

    yield load_data_and_test_cases

    session.close()
    Base.metadata.drop_all(engine)
    

@pytest_asyncio.fixture(scope="function")
async def create_async_data_driven_db(request):
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    
    engine = create_async_engine("sqlite+aiosqlite:///:memory:")

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    async_session_factory = async_sessionmaker(engine, expire_on_commit=False)

    async def load_data_and_provide_session(json_file_name: str, type: str = "integration"):
        resource_path = RESOURCE_PATH / type / json_file_name
        if not resource_path.exists():
            raise FileNotFoundError(f"Could not find fixture file: {resource_path}")

        with open(resource_path, 'r') as f:
            data = json.load(f)

        async with async_session_factory() as session:
            setup_data = data.get("setup_data", [])
            for item in setup_data:
                model_class = SCHEMA_MAP.get(item["model"])
                
                instance_data = item["data"].copy()
                instance_id = instance_data.pop('id', None)
                instance = model_class(**instance_data)
                if instance_id is not None:
                    instance.id = instance_id
                    
                session.add(instance)
            await session.commit()

        test_cases = data.get("test_cases", [])
        return async_session_factory, test_cases

    yield load_data_and_provide_session

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)

    await engine.dispose()
    

async def _load_and_populate_db(engine, json_file_path):
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    async_session_factory = async_sessionmaker(engine, expire_on_commit=False)
    
    with open(json_file_path, 'r') as f:
        data = json.load(f)

    async with async_session_factory() as session:
        setup_data = data.get("setup_data", [])
        for item in setup_data:
            model_class = SCHEMA_MAP.get(item["model"])
            instance_data = item["data"].copy()
            instance_id = instance_data.pop('id', None)
            instance = model_class(**instance_data)
            if instance_id is not None:
                instance.id = instance_id
            session.add(instance)
        await session.commit()
    
    return async_session_factory

@pytest_asyncio.fixture(scope="function")
async def prepared_session_factory(request):
    """ _summary_

        Raises:
            FileNotFoundError: _description_

        Returns:
            Any: _description_
    """
    json_file_path = request.param
    engine = create_async_engine("sqlite+aiosqlite:///:memory:")
    
    session_factory = await _load_and_populate_db(engine, json_file_path)

    yield session_factory

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
    await engine.dispose()


def load_all_assertions():
    
    pytest_params = []
    for test_file in TEST_FILES:
       
        test_type = test_file.parent.name
        marker = getattr(pytest.mark, test_type, None)

        try:
            with open(test_file, 'r', encoding='utf-8') as f:
                data = json.load(f)
        except (json.JSONDecodeError, FileNotFoundError):
            logger.warning(f"Could not read or decode JSON from: {test_file}. Skipping.")
            continue

        for case in data.get("test_cases", []):
            case_desc = case.get("description", "unnamed_case")
            for i, assertion in enumerate(case.get("assertions", [])):
                assertion_desc = assertion.get('params', {}).get('attribute', assertion['type'])
                test_id = f"{test_file.name}-{case_desc}[{i}]-{assertion_desc}"

                param = pytest.param(
                    test_file,          
                    assertion,          
                    marks=[marker] if marker else [],
                    id=test_id        
                )
                pytest_params.append(param)


    return pytest_params


def pytest_generate_tests(metafunc):
    if "prepared_session_factory" in metafunc.fixturenames and "assertion_data" in metafunc.fixturenames:
        pytest_params = load_all_assertions()
        metafunc.parametrize(
            "prepared_session_factory, assertion_data",
            pytest_params, 
            indirect=["prepared_session_factory"]
        )