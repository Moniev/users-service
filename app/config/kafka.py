from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
import asyncio, ssl
from .settings import settings
from loguru import logger
from typing import Any, Dict, Optional


def create_kafka_ssl_context() -> Optional[ssl.SSLContext]:
    if settings.KAFKA_SECURITY_PROTOCOL not in ("SSL", "SASL_SSL"):
        logger.info("Kafka security protocol does not require SSL. Skipping SSL context creation.")
        return None

    try:
        if not all([settings.KAFKA_SSL_CA_PATH, settings.KAFKA_SSL_CERT_PATH, settings.KAFKA_SSL_KEY_PATH]):
            logger.error("Kafka SSL protocol is configured, but one or more certificate paths (CA, CERT, KEY) are missing.")
            raise ValueError("For SSL protocol, KAFKA_SSL_CA_PATH, KAFKA_SSL_CERT_PATH, and KAFKA_SSL_KEY_PATH are mandatory.")

        context = ssl.create_default_context(
            purpose=ssl.Purpose.SERVER_AUTH,
            cafile=settings.KAFKA_SSL_CA_PATH
        )

        context.load_cert_chain(
            certfile=settings.KAFKA_SSL_CERT_PATH,
            keyfile=settings.KAFKA_SSL_KEY_PATH
        )
        
        logger.info("Kafka SSL context with client certificate created successfully.")
        return context

    except FileNotFoundError as e:
        logger.error(f"A Kafka certificate file was not found: {e}")
        raise RuntimeError(f"Could not create Kafka SSL context due to a missing file: {e}")
    except Exception as e:
        logger.error(f"An unexpected error occurred while creating Kafka SSL context: {e}")
        raise RuntimeError(f"Failed to create Kafka SSL context: {e}")


kafka_ssl_context = create_kafka_ssl_context()
kafka_connection_params: Dict[str, Any] = {
    "bootstrap_servers": settings.KAFKA_BOOTSTRAP_SERVERS,
    "client_id": settings.KAFKA_CLIENT_ID,
    "security_protocol": settings.KAFKA_SECURITY_PROTOCOL,
    "ssl_context": kafka_ssl_context,
}


kafka_producer: AIOKafkaProducer = AIOKafkaProducer(**kafka_connection_params)


async def get_kafka_producer() -> AIOKafkaProducer:
    return kafka_producer


async def consume_messages(topic: str) -> None:
    consumer_params = {
        **kafka_connection_params, 
        "group_id": f"{settings.KAFKA_CLIENT_ID}-{topic}-group", 
        "auto_offset_reset": 'earliest',
    }
    
    consumer = AIOKafkaConsumer(topic, **consumer_params)
    
    try:
        await consumer.start()
        logger.info(f"Kafka consumer started for topic '{topic}'...")
        async for msg in consumer:
            logger.info(
                f"Consumed message from topic {msg.topic}: "
                f"key={msg.key.decode() if msg.key else ''}, "
                f"value={msg.value.decode() if msg.value else ''}"
            )
    except asyncio.CancelledError:
        logger.info(f"Stopping Kafka consumer for topic '{topic}' due to cancellation.")
    except Exception as e:
        logger.error(f"Error consuming messages from topic '{topic}': {e}")
    finally:
        logger.info(f"Kafka consumer for topic '{topic}' stopped.")
        await consumer.stop()

consumer_tasks = [
    asyncio.create_task(consume_messages("users-events")),
    asyncio.create_task(consume_messages("auth-events")),
]
