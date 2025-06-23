import asyncio, uvicorn
from app.controllers import auth_controller, users_controller, diagnostics_controller
from app.config.database import create_db_and_tables, engine
from app.config.kafka import kafka_producer, consumer_tasks
from app.config.logger import setup_logger
from app.config.redis import redis_client, close_redis_connection
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger
from prometheus_fastapi_instrumentator import Instrumentator
from prometheus_client import ProcessCollector

setup_logger()


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Application startup...")
    logger.info("Connecting to database and other services...")

    try:
        await create_db_and_tables()
    except Exception as e:
        logger.error(f"Could not connect to database: {e}")
    
    for task in consumer_tasks:
        logger.info(f"Starting consumer task: {task.get_name()}")
    
    try:
        await redis_client.ping()
        logger.info("Successfully connected to Redis.")
    except Exception as e:
        logger.error(f"Could not connect to Redis: {e}")
    
    logger.info("Starting Kafka producer...")
    await kafka_producer.start()
    
    instrumentator.expose(app)
    logger.info("Prometheus instrumentator exposed on /metrics endpoint.")
    
    yield 

    for task in consumer_tasks:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            logger.info(f"Consumer task {task.get_name()} cancelled successfully.")

    logger.info("Stopping Kafka producer...")
    await kafka_producer.stop()

    await engine.dispose()
    await close_redis_connection()

    logger.info("Application shutdown...")
    logger.info("Closing database connections...")


app: FastAPI = FastAPI(
    title="Factory Chainlines - Users Service",
    description="API for managing users, authentication, and system diagnostics.",
    version="1.0.0",
    contact={
        "name": "Robert Moń",
        "email": "m0ni3v@gmail.com",
    },
    license_info={
        "name": "Apache 2.0",
        "url": "https://www.apache.org/licenses/LICENSE-2.0.html",
    },
    lifespan=lifespan
)

instrumentator: Instrumentator = Instrumentator().instrument(app)
ProcessCollector(namespace="users-service")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost", "http://localhost:3000"], 
    allow_credentials=True,
    allow_methods=["*"], 
    allow_headers=["*"], 
)

app.include_router(auth_controller.router, prefix="/api/v1")
app.include_router(diagnostics_controller.router, prefix="/api/v1")
app.include_router(users_controller.router, prefix="/api/v1")

@app.get("/", tags=["Root"])
def read_root():
    return {"status": "ok", "message": "Welcome to the Factory Chainlines Users Service API"}


if __name__ == "__main__":
    logger.info("Starting development server with Uvicorn...")
    uvicorn.run(
        "app.main:app", 
        host="0.0.0.0", 
        port=8000,
        reload=True,     
        log_level="info"
    )