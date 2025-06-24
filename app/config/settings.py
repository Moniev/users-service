from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import computed_field, PostgresDsn
from typing import Optional

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

    DB_USER: Optional[str] = None 
    DB_PASSWORD: Optional[str] = None
    DB_HOST: Optional[str] = "postgres.postgres.svc.cluster.local"
    DB_PORT: Optional[int] = 5432
    DB_NAME: Optional[str] = "MAIN_DB" 
    
    @computed_field
    @property
    def DATABASE_URI(self) -> str:
        return str(PostgresDsn.build(
            scheme="postgresql+asyncpg",
            username=self.DB_USER,
            password=self.DB_PASSWORD,
            host=self.DB_HOST,
            port=self.DB_PORT,
            path=f"{self.DB_NAME or ''}",
        ))

    DB_SSL_CA_PATH: Optional[str] = "/etc/users-service-tls/postgres/ca.crt"
    DB_SSL_CERT_PATH: Optional[str] = "/etc/users-service-tls/postgres/tls.crt"
    DB_SSL_KEY_PATH: Optional[str] = "/etc/users-service-tls/postgres/tls.key"
    
    GOOGLE_CLIENT_ID: Optional[str] = None
    GOOGLE_CLIENT_SECRET: Optional[str] = None
    
    KAFKA_BOOTSTRAP_SERVERS: str = "kafka.kafka.svc.cluster.local:9093"
    KAFKA_CLIENT_ID: str = "users-service"
    KAFKA_SECURITY_PROTOCOL: str = "SSL"
    KAFKA_SASL_MECHANISM: Optional[str] = None      
    KAFKA_SASL_USERNAME: Optional[str] = None
    KAFKA_SASL_PASSWORD: Optional[str] = None
    KAFKA_SSL_CA_PATH: Optional[str] = "/etc/users-service-tls/kafka/ca.crt"
    KAFKA_SSL_CERT_PATH: Optional[str] = "/etc/users-service-tls/kafka/tls.crt"
    KAFKA_SSL_KEY_PATH: Optional[str] = "/etc/users-service-tls/kafka/tls.key"
    
    REDIS_HOST: str = "redis.redis.svc.cluster.local"
    REDIS_PORT: int = 6379
    REDIS_DB: Optional[int] = 0
    REDIS_PASSWORD: Optional[str] = None 
    REDIS_SSL_CA_PATH: Optional[str] = "/etc/users-service-tls/redis/ca.crt"
    REDIS_SSL_CERT_PATH: Optional[str] = "/etc/users-service-tls/redis/tls.crt"
    REDIS_SSL_KEY_PATH: Optional[str] = "/etc/users-service-tls/redis/tls.key"
    
    @computed_field
    @property
    def REDIS_URI(self) -> str:
        scheme = "rediss" if self.REDIS_SSL_CA_PATH else "redis"
        auth = f":{self.REDIS_PASSWORD}@" if self.REDIS_PASSWORD else ""
        
        return f"{scheme}://{auth}{self.REDIS_HOST}:{self.REDIS_PORT}/{self.REDIS_DB}"

settings: Settings = Settings()