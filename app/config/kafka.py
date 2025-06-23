import asyncio
from loguru import logger
from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
from .settings import settings


kafka_producer: AIOKafkaProducer = AIOKafkaProducer(
    bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
    client_id=settings.KAFKA_CLIENT_ID,
    security_protocol=settings.KAFKA_SECURITY_PROTOCOL,
    sasl_mechanism=settings.KAFKA_SASL_MECHANISM,
    sasl_plain_username=settings.KAFKA_SASL_USERNAME,
    sasl_plain_password=settings.KAFKA_SASL_PASSWORD,
    ssl_context=None 
)

async def get_kafka_producer() -> AIOKafkaProducer:
    return kafka_producer

async def consume_messages(topic: str):
    consumer = AIOKafkaConsumer(
        topic,
        bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
        client_id=f"{settings.KAFKA_CLIENT_ID}-consumer",
        group_id=f"{settings.KAFKA_CLIENT_ID}-group",
        auto_offset_reset='earliest',
        security_protocol=settings.KAFKA_SECURITY_PROTOCOL,
        sasl_mechanism=settings.KAFKA_SASL_MECHANISM,
        sasl_plain_username=settings.KAFKA_SASL_USERNAME,
        sasl_plain_password=settings.KAFKA_SASL_PASSWORD,
        ssl_context=None
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
    finally:
        logger.info(f"Stopping Kafka consumer for topic '{topic}'...")
        await consumer.stop()


consumer_tasks = [
    asyncio.create_task(consume_messages("users-events")),
    asyncio.create_task(consume_messages("auth-events")),
]