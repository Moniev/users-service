from app.config.database import Base
from app.models.schema.activation_code import ActivationCode
from app.models.schema.second_factor_code import SecondFactorCode
from app.models.schema.reset_code import ResetCode
from app.models.schema.user import User, password_hasher
from app.models.schema.user_device import UserDevice
from app.models.schema.user_role import UserRole
from app.models.schema.user_settings import UserSettings
from app.models.schema.verification_code import VerificationCode
from app.config.settings import Settings
from app.config.logger import setup_logger
from app.config.kafka import get_kafka_producer, stop_kafka_producer 
from app.services.auth_service import AuthService
from app.services.users_service import UsersService
from app.services.diagnostics_service import DiagnosticsService
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
import fakeredis.aioredis as fakeredis
import json, pytest, pytest_asyncio, yaml, os
from loguru import logger
from pathlib import Path
from pydantic import BaseModel 
from redis import Redis, ConnectionPool
import redis.asyncio as redis
from sqlalchemy import create_engine, select, func
from sqlalchemy.orm import Session, selectinload, joinedload
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from testcontainers.redis import RedisContainer
from testcontainers.kafka import KafkaContainer
from types import SimpleNamespace
from typing import Any, Generator, Callable, Tuple, Dict, List
from typing import AsyncGenerator, Generator, Callable, Tuple, List, Dict, Any, Optional
from unittest.mock import AsyncMock, patch


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
    "user_get_user_by_activation_code": User.get_user_by_activation_code,
    "user_get_user_by_verification_code": User.get_user_by_verification_code,
    "user_get_user_by_second_factor_code": User.get_user_by_second_factor_code,
    "user_get_user_by_reset_code": User.get_user_by_reset_code,
    "activate_account": User.activate_account,
    "verify_account": User.verify_account
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

SERVICE_MAP: dict[str, Any] = {
    "AuthService": AuthService,
    "DiagnosticsService": DiagnosticsService,
    "UsersService": UsersService,
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


TEST_TYPE_PATHS: dict[str, Path] = {
    "unit": UNIT_TEST_PATH,
    "integration": INTEGRATION_TEST_PATH,
    "e2e": E2E_TEST_PATH,
}

setup_logger(LOGS_PATH)


def pytest_configure(config):
    config.addinivalue_line("markers", "integration: mark tests as integration tests")
    config.addinivalue_line("markers", "unit: mark tests as unit tests")
    config.addinivalue_line("markers", "e2e: mark tests as end-to-end tests")


@pytest.fixture(scope="function")
def test_environment(request):
    env = SimpleNamespace()

    if request.node.get_closest_marker("unit"):
        logger.info("Test-type: UNIT. Providing MOCKED environment.")
        env.redis_client = request.getfixturevalue("mock_redis_client_session")
        env.kafka_producer = request.getfixturevalue("mock_kafka_producer_session")
        env.settings = request.getfixturevalue("app_settings")

    else: 
        logger.info("Test-type: INTEGRATION/E2E. Providing REAL environment with Testcontainers.")
        env.redis_client = request.getfixturevalue("redis_client_session")
        env.kafka_producer = request.getfixturevalue("kafka_producer_client")
        env.settings = request.getfixturevalue("test_app_settings") 

    return env


@pytest.fixture(scope="session")
def jwt_key_paths(tmp_path_factory) -> Dict[str, str]:
    key_dir = tmp_path_factory.mktemp("jwt_keys")
    private_key_path = key_dir / "jwt_private.pem"
    public_key_path = key_dir / "jwt_public.pem"

    private_key = rsa.generate_private_key(
        public_exponent=65537,
        key_size=2048,
    )

    pem_private = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )
    with open(private_key_path, "wb") as f:
        f.write(pem_private)

    public_key = private_key.public_key()
    pem_public = public_key.public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo,
    )
    with open(public_key_path, "wb") as f:
        f.write(pem_public)
        
    logger.info(f"Wygenerowano tymczasowe klucze JWT w katalogu: {key_dir}")

    return {
        "private": str(private_key_path),
        "public": str(public_key_path),
    }


@pytest_asyncio.fixture(scope="session")
async def kafka_producer_session():
    producer = await get_kafka_producer()
    yield producer
    await stop_kafka_producer()


@pytest_asyncio.fixture(scope="session")
async def redis_container_info() -> AsyncGenerator[Tuple[str, int], None]:
    with RedisContainer("redis:7-alpine") as redis_container:
        redis_host = redis_container.get_container_host_ip()
        redis_port = redis_container.get_exposed_port(6379)

        logger.info(f"Testcontainers: Redis container running at {redis_host}:{redis_port}")

        yield redis_host, int(redis_port) 


@pytest_asyncio.fixture(scope="session")
async def kafka_container_info() -> AsyncGenerator[str, None]:
    with KafkaContainer("confluentinc/cp-kafka:7.5.0") as kafka_container:
        kafka_bootstrap_servers = kafka_container.get_bootstrap_server()
        logger.info(f"Testcontainers: Kafka container running, bootstrap servers: {kafka_bootstrap_servers}")
        yield kafka_bootstrap_servers


@pytest.fixture(scope="session")
def test_app_settings(
    app_settings: Settings,  
    redis_container_info: Tuple[str, int], 
    kafka_container_info: str, 
    jwt_key_paths: Dict[str, str]
) -> Generator[Settings, Any, Any]:
    
    settings_for_test = app_settings
    
    redis_host, redis_port = redis_container_info
    settings_for_test.REDIS_HOST = redis_host
    settings_for_test.REDIS_PORT = redis_port
    settings_for_test.REDIS_SSL_CA_PATH = "" 
    settings_for_test.REDIS_SSL_CERT_PATH = ""
    settings_for_test.REDIS_SSL_KEY_PATH = ""
    settings_for_test.REDIS_DB = 0
    settings_for_test.REDIS_PASSWORD = None

    settings_for_test.KAFKA_BOOTSTRAP_SERVERS = kafka_container_info
    settings_for_test.KAFKA_SECURITY_PROTOCOL = "PLAINTEXT" 
    settings_for_test.KAFKA_SASL_MECHANISM = None
    settings_for_test.KAFKA_SASL_USERNAME = None
    settings_for_test.KAFKA_SASL_PASSWORD = None
    settings_for_test.KAFKA_SSL_CA_PATH = ""
    settings_for_test.KAFKA_SSL_CERT_PATH = ""
    settings_for_test.KAFKA_SSL_KEY_PATH = ""
    
    settings_for_test.JWT_PRIVATE_KEY_PATH = jwt_key_paths["private"]
    settings_for_test.JWT_PUBLIC_KEY_PATH = jwt_key_paths["public"]
    
    logger.info("Conftest: Test Settings Object created from settings.yaml and overridden by test containers.")
    logger.info(f"Conftest: Redis URI -> {settings_for_test.REDIS_URI}")
    logger.info(f"Conftest: Kafka Broker -> {settings_for_test.KAFKA_BOOTSTRAP_SERVERS}")

    yield settings_for_test


@pytest.fixture(scope="session")
def app_settings() -> Settings:
    if not SETTINGS_PATH.exists():
        raise FileNotFoundError(f"Failed to find settings.yaml in path {SETTINGS_PATH}")
    
    with open(SETTINGS_PATH, 'r') as f:
        yaml_data = yaml.safe_load(f)
        if not yaml_data:
            raise ValueError("settings.yaml is empty or invalid.")
    
    settings: Settings = Settings.model_validate(yaml_data)
    
    os.environ["DOCKER_HOST"] = settings.DOCKER_HOST
    if os.environ.get("CI") == "true":
        logger.info("CI environment detected. Forcing DOCKER_HOST to None to use runner's default.")
        settings.DOCKER_HOST = None 
    
    os.environ["TEST_MODE"] = "True"
    
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
            
            if item["model"] == "User" and "password" in instance_data:
                instance_data["password"] = password_hasher.hash(instance_data["password"])
            
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
    

@pytest.fixture(scope="function")
async def create_async_data_driven_db(request) -> AsyncGenerator[Callable[[str, str], Tuple[async_sessionmaker[AsyncSession], List[Dict]]], None]:
    engine = create_async_engine("sqlite+aiosqlite:///:memory:", echo=True)

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    async_session_factory = async_sessionmaker(engine, expire_on_commit=False)

    async def load_data_and_provide_session(json_file_name: str, type: str = "integration") -> Tuple[async_sessionmaker[AsyncSession], List[Dict]]:
        resource_path = RESOURCE_PATH / type / json_file_name
        if not resource_path.exists():
            raise FileNotFoundError(f"Could not find fixture file: {resource_path}")

        with open(resource_path, 'r') as f:
            data = json.load(f)

        async with async_session_factory() as session:
            setup_data = data.get("setup_data", [])
            for item in setup_data:
                model_class = SCHEMA_MAP.get(item["model"])
                if not model_class:
                    raise ValueError(f"Unknown model: '{item['model']}' in JSON file.")
                
                instance_data: Dict[str, Any] = item["data"].copy()
                
                if item["model"] == "User" and "password" in instance_data:
                    instance_data["password"] = password_hasher.hash(instance_data["password"])

                instance_id: Optional[int] = instance_data.pop('id', None)
                instance: Base = model_class(**instance_data)
                
                if instance_id is not None:
                    instance.id = instance_id
                    
                session.add(instance)
            
            try:
                await session.commit()
            except Exception as e:
                await session.rollback()
                logger.error(f"Error during setup data commit: {e}")
                raise

        test_cases = data.get("test_cases", [])
        return async_session_factory, test_cases 

    yield load_data_and_provide_session

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)

    await engine.dispose()
    

async def load_and_populate_db(engine, json_file_path):
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
async def prepared_session_factory(request) -> AsyncGenerator[async_sessionmaker[AsyncSession], None]:
    json_file_path = request.param 
    engine = create_async_engine("sqlite+aiosqlite:///:memory:", echo=False)

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all) 

    async_session_factory = async_sessionmaker(engine, expire_on_commit=False)
    
    if json_file_path: 
        with open(json_file_path, 'r') as f:
            data = json.load(f)

        async with async_session_factory() as session:
            setup_data = data.get("setup_data", [])
            for item in setup_data:
                model_class = SCHEMA_MAP.get(item["model"])
                if not model_class:
                    raise ValueError(f"Unknown model: '{item['model']}' in JSON file.")
                
                instance_data = item["data"].copy()
                if item["model"] == "User" and "password" in instance_data:
                    instance_data["password"] = password_hasher.hash(instance_data["password"])
                
                instance_id = instance_data.pop('id', None)
                instance = model_class(**instance_data)
                if instance_id is not None:
                    instance.id = instance_id
                session.add(instance)
            
            try:
                await session.commit()
            except Exception as e:
                await session.rollback()
                logger.error(f"Error during setup data commit for {json_file_path}: {e}")
                raise

    yield async_session_factory 

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
    await engine.dispose()


@pytest_asyncio.fixture(scope="function")
async def mock_redis_client_session() -> AsyncGenerator[Redis, Any]:
    logger.info("Creating in-memory Redis client (fakeredis) for test session.")
    client = fakeredis.FakeRedis()
    try:
        await client.ping()
        logger.success("In-memory Redis (fakeredis) client created successfully!")
    except Exception as e:
        logger.error(f"Failed to create fakeredis client: {e}")
        pytest.fail(f"Failed to create fakeredis client: {e}")

    yield client

    logger.info("Closing in-memory Redis (fakeredis) client.")
    await client.close()


@pytest_asyncio.fixture(scope="function")
async def redis_client_session(test_app_settings: Settings) -> AsyncGenerator[redis.Redis, None]: 
    logger.info(f"Creating Redis connection pool for test session using settings: {test_app_settings.REDIS_URI}")
    
    pool_kwargs = {
        "max_connections": 10,
        "decode_responses": True
    }

    pool: ConnectionPool = redis.ConnectionPool.from_url(
        test_app_settings.REDIS_URI,
        **pool_kwargs
    )
    client: redis.Redis = redis.Redis(connection_pool=pool)

    try:
        await client.ping()
        logger.success("Redis client successfully connected and pinged!")
        yield client
    except Exception as e:
        logger.error(f"Failed to connect to Redis during test setup: {e}")
        pytest.fail(f"Failed to connect to Redis: {e}")
    finally:
        logger.info("Closing Redis connection from test fixture.")
        await client.close()
        await client.connection_pool.disconnect()


@pytest_asyncio.fixture(scope="function")
async def mock_kafka_producer_session():
    logger.info("Creating mock Kafka producer for test session.")
    mock_producer = AsyncMock()
    mock_producer.send.return_value = AsyncMock() 
    yield mock_producer
    logger.info("Mock Kafka producer closed.")


@pytest_asyncio.fixture(scope="function")
async def kafka_producer_client(test_app_settings: Settings) -> AsyncGenerator[Any, None]:
    producer = await get_kafka_producer(settings=test_app_settings)
    try:
        logger.success("Kafka producer client created and ready!")
        yield producer
    except Exception as e:
        logger.error(f"Failed to start Kafka producer client for tests: {e}")
        pytest.fail(f"Could not start Kafka producer client: {e}")
    finally:
        logger.info("Stopping Kafka producer client after test session.")
        await stop_kafka_producer()


@pytest_asyncio.fixture(scope="session")
async def redis_server() -> AsyncGenerator[Redis, None]:
    with RedisContainer("redis:7-alpine") as redis_container:
        redis_url = redis_container.get_container_host_ip()
        redis_port = redis_container.get_exposed_port(6379) 

        logger.info(f"Testcontainers: Redis container running at {redis_url}:{redis_port}")

        client = redis.Redis(host=redis_url, port=redis_port, decode_responses=True)

        try:
            await client.ping()
            logger.success("Testcontainers: Redis client successfully connected and pinged!")
            yield client
        except Exception as e:
            logger.error(f"Testcontainers: Failed to connect or ping Redis: {e}")
            pytest.fail(f"Could not connect to Redis container: {e}")
        finally:
            logger.info("Testcontainers: Closing Redis connection from test fixture.")
            await client.close()
            await client.connection_pool.disconnect()


@pytest_asyncio.fixture(scope="session")
async def kafka_producer_session(test_app_settings: Settings) -> AsyncGenerator[Any, None]: 
    with KafkaContainer("confluentinc/cp-kafka:7.5.0") as kafka_container:
        kafka_bootstrap_servers = kafka_container.get_bootstrap_server()
        logger.info(f"Testcontainers: Kafka container running, bootstrap servers: {kafka_bootstrap_servers}")

        test_app_settings.KAFKA_BOOTSTRAP_SERVERS = kafka_bootstrap_servers
        test_app_settings.KAFKA_SECURITY_PROTOCOL = "PLAINTEXT"
        test_app_settings.KAFKA_SASL_MECHANISM = None
        test_app_settings.KAFKA_SASL_USERNAME = None
        test_app_settings.KAFKA_SASL_PASSWORD = None
        test_app_settings.KAFKA_SSL_CA_PATH = ""
        test_app_settings.KAFKA_SSL_CERT_PATH = ""
        test_app_settings.KAFKA_SSL_KEY_PATH = ""
        

        from app.config.kafka import get_kafka_producer, stop_kafka_producer

        producer = await get_kafka_producer(settings=test_app_settings) 
        try:
            logger.success("Testcontainers: Kafka producer created and ready!")
            yield producer
        except Exception as e:
            logger.error(f"Testcontainers: Failed to start Kafka producer for tests: {e}")
            pytest.fail(f"Could not start Kafka producer: {e}")
        finally:
            logger.info("Testcontainers: Stopping Kafka producer after test session.")
            await stop_kafka_producer()


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


def load_all_test_cases():
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
            test_id = f"{test_file.name}-{case_desc}"

            param = pytest.param(
                test_file,
                case,  
                marks=[marker] if marker else [],
                id=test_id
            )
            pytest_params.append(param)
    return pytest_params


def pytest_generate_tests(metafunc):
    if "prepared_session_factory" not in metafunc.fixturenames or "test_case_data" not in metafunc.fixturenames:
        return

    pytest_params = []
    
    marker_expression = metafunc.config.getoption("-m")

    paths_to_scan = []

    if not marker_expression:
        logger.info("No marker provided. Scanning all test directories.")
        paths_to_scan.extend(TEST_TYPE_PATHS.values())
    else:
        logger.info(f"Marker expression '{marker_expression}' provided. Scanning specific directories.")
        for marker, path in TEST_TYPE_PATHS.items():
            if marker in marker_expression:
                paths_to_scan.append(path)
    
    if not paths_to_scan:
        logger.warning(f"Could not match marker expression '{marker_expression}' to any known test path. Scanning all paths as a fallback.")
        paths_to_scan.extend(TEST_TYPE_PATHS.values())


    for path in paths_to_scan:
        test_type = path.name
        marker = getattr(pytest.mark, test_type, None)

        for test_file in path.glob("*.json"):
            try:
                with open(test_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
            except (json.JSONDecodeError, FileNotFoundError):
                logger.warning(f"Could not read or decode JSON from: {test_file}. Skipping.")
                continue

            for case in data.get("test_cases", []):
                case_desc = case.get("description", "unnamed_case")
                test_id = f"{test_file.name}-{case_desc}"

                param = pytest.param(
                    test_file,
                    case,
                    marks=[marker] if marker else [],
                    id=test_id
                )
                pytest_params.append(param)

    metafunc.parametrize(
        "prepared_session_factory,test_case_data",
        pytest_params,
        indirect=["prepared_session_factory"]
    )
        
        
SERVICE_DEPENDENCIES: dict[object, list[str]] = {
    AuthService: [
        "session_factory",
        "redis_client",
        "kafka_producer",
        "settings"
    ],
    DiagnosticsService: [],
    UsersService: ["session_factory"] 
}