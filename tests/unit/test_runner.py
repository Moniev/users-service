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
        if hasattr(model_class, relationship_name) and relationship_name in model_class.__mapper__.relationships:
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
            
            if "filter_by" in params:
                stmt = stmt.filter_by(**params["filter_by"])
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
            
        elif assertion_type == "is_not_none":
            instance = await get_object(session, params)
            assert instance is not None

        elif assertion_type == "execute_function":
            func_name = params["function_name"]
            target_function = FUNCTION_MAP.get(func_name)
            if not target_function:
                pytest.fail(f"Function '{func_name}' not found in FUNCTION_MAP.")
            
            func_params = params.get("function_params", {})
            call_params = {"async_session": session_factory}

            if "on_instance" in func_params:
                instance_config = func_params.pop("on_instance")
                model_class = SCHEMA_MAP.get(instance_config["model"])
                instance = await session.get(model_class, instance_config["lookup"]["id"])
                assert instance is not None, f"Instance {instance_config['model']} with lookup {instance_config['lookup']} not found."
                
                method_to_call = getattr(instance, func_name.split('.')[-1])
                
                method_params = func_params.get("method_params", {})
                final_method_args = {}
                for name, value in method_params.items():
                    if isinstance(value, dict) and "create_dummy" in value:
                        dummy_config = value["create_dummy"]
                        dummy_model = SCHEMA_MAP[dummy_config["model"]]
                        dummy_data = dummy_config["data"].copy() 
                        
                        parent_model_name = instance_config["model"].lower()
                        foreign_key_name = f"{parent_model_name}_id"
                        if foreign_key_name not in dummy_data:
                            dummy_data[foreign_key_name] = instance.id
                        
                        final_method_args[name] = dummy_model(**dummy_data)
                    else:
                        final_method_args[name] = value

                result = await method_to_call(**call_params, **final_method_args)

            else:
                 target_function = FUNCTION_MAP.get(func_name)
                 if not target_function:
                    pytest.fail(f"Function '{func_name}' not found in FUNCTION_MAP.")
                 
                 call_params.update(func_params)
                 if inspect.iscoroutinefunction(target_function):
                    result = await target_function(**call_params)
                 else:
                    result = target_function(**call_params) 
            
            expected = params["expected_result"]
            op = expected["operator"]
            if op == "is_not_none":
                assert result is not None, f"Expected result to not be None, but it was."
            elif op == "is_none":
                assert result is None, f"Expected result to be None, but it was not."
            elif op == "equals":
                assert result == expected["value"]
            else:
                pytest.fail(f"Unknown result operator: '{op}'")
        
        else:
            pytest.fail(f"Unknown assertion type: '{assertion_type}'")
