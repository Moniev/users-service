from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
import asyncio
import os, ssl
from loguru import logger
from typing import Any, Dict, List, Optional
from .settings import Settings 



kafka_producer_instance: Optional[AIOKafkaProducer] = None 
consumer_tasks: List[asyncio.Task] = [] 


def create_kafka_ssl_context(settings: Settings) -> Optional[ssl.SSLContext]:
    if not all([settings.KAFKA_SSL_CA_PATH, settings.KAFKA_SSL_CERT_PATH, settings.KAFKA_SSL_KEY_PATH]):
        logger.info("Kafka SSL certificate paths are not fully configured or are empty. Skipping SSL context creation.")
        return None

    if settings.KAFKA_SECURITY_PROTOCOL not in ("SSL", "SASL_SSL"):
        logger.info("Kafka security protocol does not require SSL. Skipping SSL context creation.")
        return None

    try:
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
        logger.error(f"A Kafka certificate file was not found: {e}. Path: {e.filename}")
        raise RuntimeError(f"Could not create Kafka SSL context due to a missing file: {e}")
    except Exception as e:
        logger.error(f"An unexpected error occurred while creating Kafka SSL context: {e}")
        raise RuntimeError(f"Failed to create Kafka SSL context: {e}")


async def get_kafka_producer(settings: Settings) -> Optional[AIOKafkaProducer]:
    global kafka_producer_instance 
    if kafka_producer_instance is not None:
        logger.info("Returning existing Kafka producer instance.") 
        return kafka_producer_instance

    logger.info("Initializing Kafka producer...")

    kafka_connection_params: Dict[str, Any] = {
        "bootstrap_servers": settings.KAFKA_BOOTSTRAP_SERVERS,
        "client_id": settings.KAFKA_CLIENT_ID,
        "security_protocol": settings.KAFKA_SECURITY_PROTOCOL,
    }

    if settings.KAFKA_SASL_MECHANISM:
        kafka_connection_params["sasl_mechanism"] = settings.KAFKA_SASL_MECHANISM
        kafka_connection_params["sasl_plain_username"] = settings.KAFKA_SASL_USERNAME
        kafka_connection_params["sasl_plain_password"] = settings.KAFKA_SASL_PASSWORD

    ssl_context = create_kafka_ssl_context(settings) 
    if ssl_context:
        kafka_connection_params["ssl_context"] = ssl_context
    elif settings.KAFKA_SECURITY_PROTOCOL in ("SSL", "SASL_SSL"):
        pass

    try:
        kafka_producer_instance = AIOKafkaProducer(**kafka_connection_params, request_timeout_ms=5000) 
        await kafka_producer_instance.start()
        logger.info(f"Kafka producer started and connected to {kafka_connection_params['bootstrap_servers']}.")
    except Exception as e:
        logger.error(f"Failed to create and start Kafka producer: {e}. Params: {kafka_connection_params}")
        raise  

    return kafka_producer_instance


async def stop_kafka_producer():
    global kafka_producer_instance 
    if kafka_producer_instance:
        logger.info("Stopping Kafka producer...")
        await kafka_producer_instance.stop() 
        kafka_producer_instance = None 
        logger.info("Kafka producer stopped.")



async def consume_messages(topic: str, settings: Settings) -> None: 
    consumer_params = {
        "bootstrap_servers": settings.KAFKA_BOOTSTRAP_SERVERS,
        "client_id": settings.KAFKA_CLIENT_ID,
        "security_protocol": settings.KAFKA_SECURITY_PROTOCOL,
        "group_id": f"{settings.KAFKA_CLIENT_ID}-{topic}-group",
        "auto_offset_reset": 'earliest',
    }

    if settings.KAFKA_SASL_MECHANISM:
        consumer_params["sasl_mechanism"] = settings.KAFKA_SASL_MECHANISM
        consumer_params["sasl_plain_username"] = settings.KAFKA_SASL_USERNAME
        consumer_params["sasl_plain_password"] = settings.KAFKA_SASL_PASSWORD

    ssl_context = create_kafka_ssl_context(settings) 
    if ssl_context:
        consumer_params["ssl_context"] = ssl_context
    elif settings.KAFKA_SECURITY_PROTOCOL in ("SSL", "SASL_SSL"):
        pass

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


async def start_consumers(settings: Settings): 

    global consumer_tasks 
    if consumer_tasks:
        logger.warning("Consumer tasks already running.")
        return

    
    topics_to_consume = ["users-events", "auth-events"]
    for topic in topics_to_consume:
        task = asyncio.create_task(consume_messages(topic, settings)) 
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