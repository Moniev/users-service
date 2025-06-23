import asyncio, ssl
from loguru import logger
from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
from .settings import settings 


kafka_ssl_context = None
if settings.KAFKA_SECURITY_PROTOCOL in ("SSL", "SASL_SSL"):
    if settings.KAFKA_SSL_CA_PATH:
        try:
            kafka_ssl_context = ssl.create_default_context(
                purpose=ssl.Purpose.SERVER_AUTH,
                cafile=settings.KAFKA_SSL_CA_PATH
            )

            logger.info(f"Kafka SSL context created using CA path: {settings.KAFKA_SSL_CA_PATH}")
        except FileNotFoundError as e:
            logger.error(f"Kafka CA certificate file not found at {settings.KAFKA_SSL_CA_PATH}. Error: {e}")
            raise RuntimeError(f"Kafka SSL/SASL_SSL configured but CA file not found: {e}")
        except Exception as e:
            logger.error(f"Error creating Kafka SSL context: {e}")
            raise RuntimeError(f"Failed to create Kafka SSL context: {e}")
    else:
        logger.error(f"Kafka security protocol is '{settings.KAFKA_SECURITY_PROTOCOL}' but KAFKA_SSL_CA_PATH is not configured in settings.")
        raise ValueError(f"`ssl_context` is mandatory if security_protocol=='{settings.KAFKA_SECURITY_PROTOCOL}' and KAFKA_SSL_CA_PATH is missing.")
elif settings.KAFKA_SECURITY_PROTOCOL == "PLAINTEXT":
    logger.info("Kafka security protocol is PLAINTEXT. No SSL context needed.")
else:
    logger.warning(f"Unknown Kafka security protocol: {settings.KAFKA_SECURITY_PROTOCOL}")

kafka_producer_params = {
    "bootstrap_servers": settings.KAFKA_BOOTSTRAP_SERVERS,
    "client_id": settings.KAFKA_CLIENT_ID,
    "security_protocol": settings.KAFKA_SECURITY_PROTOCOL,
}

if settings.KAFKA_SECURITY_PROTOCOL in ("SASL_SSL", "SASL_PLAINTEXT"):
    if not all([settings.KAFKA_SASL_MECHANISM, settings.KAFKA_SASL_USERNAME, settings.KAFKA_SASL_PASSWORD]):
        logger.error("Kafka SASL_SSL/SASL_PLAINTEXT protocol selected, but SASL credentials are incomplete.")
        raise ValueError("SASL credentials (mechanism, username, password) are mandatory for SASL protocols.")
    
    kafka_producer_params["sasl_mechanism"] = settings.KAFKA_SASL_MECHANISM
    kafka_producer_params["sasl_plain_username"] = settings.KAFKA_SASL_USERNAME
    kafka_producer_params["sasl_plain_password"] = settings.KAFKA_SASL_PASSWORD
    
    if settings.KAFKA_SECURITY_PROTOCOL == "SASL_SSL":
        if kafka_ssl_context:
            kafka_producer_params["ssl_context"] = kafka_ssl_context
        else:
            logger.error("Kafka SASL_SSL requires SSL context, but it was not created.")
            raise RuntimeError("Kafka SASL_SSL protocol requires a valid SSL context.")
elif settings.KAFKA_SECURITY_PROTOCOL == "SSL":
    if kafka_ssl_context:
        kafka_producer_params["ssl_context"] = kafka_ssl_context
    else:
        logger.error("Kafka SSL protocol requires SSL context, but it was not created.")
        raise RuntimeError("Kafka SSL protocol requires a valid SSL context.")

kafka_producer: AIOKafkaProducer = AIOKafkaProducer(**kafka_producer_params)


async def get_kafka_producer() -> AIOKafkaProducer:
    return kafka_producer

async def consume_messages(topic: str):
    consumer_params = {
        **kafka_producer_params, 
        "group_id": f"{settings.KAFKA_CLIENT_ID}-group", 
        "auto_offset_reset": 'earliest', 
    }
    
    consumer = AIOKafkaConsumer(
        topic,
        **consumer_params
    )
    
    await consumer.start()
    logger.info(f"Kafka consumer started for topic '{topic}'...")
    try:
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