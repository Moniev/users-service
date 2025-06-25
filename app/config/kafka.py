from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
import asyncio
import os, ssl
from .settings import settings
from loguru import logger
from typing import Any, Dict, List, Optional


def create_kafka_ssl_context() -> Optional[ssl.SSLContext]:
    if os.environ.get("TEST_MODE") == "True":
        logger.warning("TEST_MODE is active. Skipping Kafka SSL context creation.")
        return None

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


kafka_producer: Optional[AIOKafkaProducer] = None


async def get_kafka_producer() -> Optional[AIOKafkaProducer]:
    global kafka_producer
    if kafka_producer is not None:
        return kafka_producer

    logger.info("Initializing Kafka producer...")

    kafka_connection_params: Dict[str, Any] = {
        "client_id": settings.KAFKA_CLIENT_ID,
    }

    if os.environ.get("TEST_MODE") == "True":
        logger.warning("TEST_MODE is active. Configuring Kafka producer for localhost.")
        kafka_connection_params["bootstrap_servers"] = "localhost:9092"
        kafka_connection_params["security_protocol"] = "PLAINTEXT"
    else:
        kafka_connection_params["bootstrap_servers"] = settings.KAFKA_BOOTSTRAP_SERVERS
        kafka_connection_params["security_protocol"] = settings.KAFKA_SECURITY_PROTOCOL
        kafka_connection_params["ssl_context"] = create_kafka_ssl_context()

    try:
        kafka_producer = AIOKafkaProducer(**kafka_connection_params, request_timeout_ms=5000)
        await kafka_producer.start()
        logger.info(f"Kafka producer started and connected to {kafka_connection_params['bootstrap_servers']}.")
    except Exception as e:
        logger.error(f"Failed to create and start Kafka producer: {e}")
        if os.environ.get("TEST_MODE") == "True":
            logger.warning("Could not connect to local Kafka. Proceeding without a producer.")
            return None
        raise  

    return kafka_producer


async def stop_kafka_producer():
    global kafka_producer
    if kafka_producer:
        logger.info("Stopping Kafka producer...")
        await kafka_producer.stop()
        kafka_producer = None
        logger.info("Kafka producer stopped.")


async def consume_messages(topic: str, kafka_connection_params: Dict[str, Any]) -> None:
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
            logger.info(f"Consumed from {msg.topic}: value={msg.value.decode() if msg.value else ''}")
    except asyncio.CancelledError:
        logger.info(f"Stopping Kafka consumer for topic '{topic}'.")
    finally:
        await consumer.stop()
        logger.info(f"Kafka consumer for topic '{topic}' has stopped.")


async def start_consumers():
    global consumer_tasks
    if consumer_tasks:
        logger.warning("Consumer tasks already running.")
        return

    if os.environ.get("TEST_MODE") == "True":
        logger.warning("TEST_MODE is active. Kafka consumers will not start.")
        return

    kafka_connection_params: Dict[str, Any] = {
        "bootstrap_servers": settings.KAFKA_BOOTSTRAP_SERVERS,
        "client_id": settings.KAFKA_CLIENT_ID,
        "security_protocol": settings.KAFKA_SECURITY_PROTOCOL,
        "ssl_context": create_kafka_ssl_context(),
    }

    topics_to_consume = ["users-events", "auth-events"]
    for topic in topics_to_consume:
        task = asyncio.create_task(consume_messages(topic, kafka_connection_params))
        consumer_tasks.append(task)
    logger.info(f"Started {len(consumer_tasks)} consumer tasks.")


async def stop_consumers():
    global consumer_tasks
    if consumer_tasks:
        logger.info(f"Stopping {len(consumer_tasks)} consumer tasks...")
        for task in consumer_tasks:
            task.cancel()
        await asyncio.gather(*consumer_tasks, return_exceptions=True)
        consumer_tasks = []
        logger.info("All consumer tasks have been stopped.")

consumer_tasks: List[asyncio.Task] = []