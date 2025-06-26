from app.utils.encrypting_utils import decrypt_data, encrypt_data
import json
from loguru import logger
from redis.asyncio.client import Redis 
from typing import Any, Dict, List, Optional
    
    
async def set_cache(obj: Any, keys: List[str], redis_client: Redis) -> bool:
    try:
        if not isinstance(redis_client, Redis):
            logger.error(f"Invalid Redis client provided to set_cache: {type(redis_client)}. Expected redis.asyncio.client.Redis")
            return False

        object_dict = obj.to_dict()
        json_string: bytes = json.dumps(object_dict).encode("utf-8")
        encrypted_object: bytes = encrypt_data(json_string)
        
        for key in keys:
            await redis_client.set(key, encrypted_object, ex=3600)
        logger.info(f"Successfully set cache for keys: {keys}")
        return True 
    except AttributeError:
        logger.error(f"Error: Object of type {type(obj)} does not have a 'to_dict()' method. Cannot serialize for cache.")
        return False
    except TypeError as e:
        logger.error(f"JSON serialization error for object {obj}: {e}. Ensure all components are JSON-serializable.")
        return False
    except Exception as e:
        logger.error(f"Unexpected error during cache setting for keys {keys}: {e}")
        return False
    

async def revoke_cache(keys: List[str], redis_client: Redis) -> bool:
    try:
        if not isinstance(redis_client, Redis):
            logger.error(f"Invalid Redis client provided to revoke_cache: {type(redis_client)}. Expected redis.asyncio.client.Redis")
            return False

        for key in keys:
            await redis_client.delete(key) 
        logger.info(f"Successfully revoked cache for keys: {keys}")
        return True 
    except Exception as e:
        logger.error(f"Unexpected error during cache deleting keys: {keys}: {e}")
        return False


async def get_cache(key: str, redis_client: Redis) -> Optional[Dict[str, Any]]:
    """
    Pobiera i deszyfruje dane z cache Redis.
    """
    if not isinstance(redis_client, Redis):
        logger.error(f"Invalid Redis client provided to get_cache: {type(redis_client)}. Expected redis.asyncio.client.Redis")
        return None

    data: Optional[bytes] = await redis_client.get(key)
    
    if not data:
        logger.debug(f"Cache miss for key: {key}")
        return None
    
    try:
        decrypted_bytes: bytes = decrypt_data(data)
        
        result_dict: Dict[str, Any] = json.loads(decrypted_bytes.decode('utf-8'))
        logger.info(f"Cache hit and successfully decrypted for key: {key}")
        return result_dict
    except ValueError as e:
        logger.error(f"Integrity check failed or decryption error for key {key}: {e}. Data might be tampered or corrupted. Deleting from cache.")
        await redis_client.delete(key) 
        return None
    except json.JSONDecodeError as e:
        logger.error(f"JSON decoding error for key {key}: {e}. Cached data might be malformed. Deleting from cache.")
        await redis_client.delete(key) 
        return None
    except Exception as e:
        logger.error(f"Unexpected error retrieving/decrypting/deserializing cache for key {key}: {e}. Deleting from cache.")
        await redis_client.delete(key) 
        return None