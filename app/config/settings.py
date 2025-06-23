from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import computed_field, AmqpDsn, PostgresDsn

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

    DB_USER: str | None = "postgres"
    DB_PASSWORD: str | None = "postgres"
    DB_HOST: str | None = "localhost"
    DB_PORT: str | None = "5432"
    DB_NAME: str | None = "postgres"
    
    @computed_field
    @property
    def DATABASE_URI(self) -> str:
        return str(PostgresDsn.build(
            scheme="postgresql+asyncpg",
            username=self.DB_USER,
            password=self.DB_PASSWORD,
            host=self.DB_HOST,
            port=self.DB_PORT,
            path=f"/{self.DB_NAME}",
        ))

    DB_SSL_CA_PATH: str | None = None
    DB_SSL_CERT_PATH: str | None = None
    DB_SSL_KEY_PATH: str | None = None
    
    GOOGLE_CLIENT_ID: str | None = None
    GOOGLE_CLIENT_SECRET: str | None = None
    
    KAFKA_BOOTSTRAP_SERVERS: str = "localhost:9092"
    KAFKA_CLIENT_ID: str = "users-service"
    KAFKA_SECURITY_PROTOCOL: str = "SASL_SSL"  
    KAFKA_SASL_MECHANISM: str | None = None      
    KAFKA_SASL_USERNAME: str | None = None
    KAFKA_SASL_PASSWORD: str | None = None
    KAFKA_SSL_CA_PATH: str | None = None
    
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 0
    REDIS_PASSWORD: str | None = None
    REDIS_SSL_CA_PATH: str | None = None
    REDIS_SSL_CERT_PATH: str | None = None
    REDIS_SSL_KEY_PATH: str | None = None
    
    @computed_field
    @property
    def REDIS_URI(self) -> str:
        scheme = "rediss" if self.REDIS_SSL_CA_PATH else "redis"
        auth = f":{self.REDIS_PASSWORD}@" if self.REDIS_PASSWORD else ""
        
        return f"{scheme}://{auth}{self.REDIS_HOST}:{self.REDIS_PORT}/{self.REDIS_DB}"

settings: Settings = Settings()