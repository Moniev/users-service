import os
import sys
from loguru import logger
from pathlib import Path

def setup_logger(log_directory: Path):
    app_env = os.getenv("APP_ENV", "dev").lower()

    logger.remove()

    log_directory.mkdir(parents=True, exist_ok=True)
    (log_directory / "log").mkdir(exist_ok=True)
    (log_directory / "json").mkdir(exist_ok=True)
    
    try:
        logger.level("AUDIT", no=25, color="<bold><magenta>", icon="🛡️")
        logger.level("PANIC", no=60, color="<bold><red on white>", icon="🚨")
    except TypeError:
        logger.trace("Loguru custom levels already configured.")

    log_configs = {
        "dev": {
            "console_level": "DEBUG",
            "file_level": "TRACE",
            "json_level": "INFO",
            "console_format": (
                "<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> | "
                "<level>{level: <8}</level> | "
                "<cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> - "
                "<level>{message}</level>"
            ),
            "console_serialize": False
        },
        "test": {
            "console_level": "INFO",
            "file_level": "DEBUG",
            "json_level": "INFO",
            "console_format": "<green>{time:HH:mm:ss}</green> | <level>{level: <8}</level> | <level>{message}</level>",
            "console_serialize": False
        },
        "prod": {
            "console_level": "INFO",
            "file_level": "INFO",
            "json_level": "INFO",
            "console_format": None, 
            "console_serialize": True 
        }
    }

    config = log_configs.get(app_env, log_configs["dev"])
    logger.info(f"Logger configured for '{app_env}' environment.")

    logger.add(
        sys.stdout, 
        level=config["console_level"],
        format=config["console_format"],
        colorize=not config["console_serialize"], 
        serialize=config["console_serialize"] 
    )

    log_file_path = log_directory / "log" / "app.log" 
    logger.add(
        log_file_path, 
        level=config["file_level"],
        rotation="100 MB",
        retention="10 days",
        compression="zip",
        format="{time} {level: <8} | {name}:{function}:{line} | {message}"
    )


    json_log_path = log_directory / "json" / "app.json"
    logger.add(
        json_log_path,
        level=config["json_level"],
        serialize=True, 
        rotation="1 week",
        retention="30 days",
        compression="zip"
    )

    logger.info("Logger configuration has been applied.")
    
