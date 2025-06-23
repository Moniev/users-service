from loguru import logger
from tests.conftest import FUNCTION_MAP, SCHEMA_MAP
import inspect, pytest
from loguru import logger
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload


async def get_object(session, params):
    model_class = SCHEMA_MAP.get(params["model"])
    if not model_class:
        pytest.fail(f"Model '{params['model']}' not found in SCHEMA_MAP.")

    lookup_criteria = params["lookup"]
    stmt = select(model_class).filter_by(**lookup_criteria)
    
    if "attribute" in params:
        relationship_name = params["attribute"].split('.')[0]
        
        if relationship_name in model_class.__mapper__.relationships:
             stmt = stmt.options(selectinload(getattr(model_class, relationship_name)))

    result = await session.execute(stmt)
    instance = result.scalar_one_or_none()
    assert instance is not None, f"Could not find instance of {params['model']} with criteria {lookup_criteria}"
    return instance


def get_nested_attribute(instance, attribute_path: str):
    current_value = instance
    for part in attribute_path.split('.'):
        current_value = getattr(current_value, part)
    return current_value


@pytest.mark.asyncio
@pytest.mark.integration
async def test_assertion_runner(prepared_session_factory, assertion_data):
    session_factory = prepared_session_factory
    assertion_type = assertion_data["type"]
    params = assertion_data["params"]
    
    logger.debug(f"--> Executing assertion: '{assertion_type}' with params {params}")

    async with session_factory() as session:
        if assertion_type == "row_count":
            model_class = SCHEMA_MAP.get(params["model"])
            stmt = select(func.count()).select_from(model_class)
            actual = (await session.execute(stmt)).scalar_one()
            assert actual == params["expected"]

        elif assertion_type == "attribute_equals":
            instance = await get_object(session, params)
            actual = get_nested_attribute(instance, params["attribute"])
            assert actual == params["expected"]
        
        elif assertion_type == "collection_has_length":
            instance = await get_object(session, params)
            actual_collection = get_nested_attribute(instance, params["attribute"])
            assert len(actual_collection) == params["expected"]

        elif assertion_type == "execute_function":
            func_name = params["function_name"]
            target_function = FUNCTION_MAP.get(func_name)
            if not target_function:
                pytest.fail(f"Function '{func_name}' not found in FUNCTION_MAP.")

            func_params = params.get("function_params", {})
            call_params = {"async_session": session_factory, **func_params}

            if inspect.iscoroutinefunction(target_function):
                result = await target_function(**call_params)
            else:
                result = target_function(**call_params) 
            
            expected = params["expected_result"]
            op = expected["operator"]
            if op == "is_not_none":
                assert result is not None
            elif op == "is_none":
                assert result is None
            elif op == "equals":
                assert result == expected["value"]
            else:
                pytest.fail(f"Unknown result operator: '{op}'")
        
        else:
            pytest.fail(f"Unknown assertion type: '{assertion_type}'")