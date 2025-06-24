import pytest
from loguru import logger
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
import inspect

from tests.conftest import FUNCTION_MAP, SCHEMA_MAP


async def get_object(session, model_name, lookup_criteria, relationships_to_load=None):
    model_class = SCHEMA_MAP.get(model_name)
    if not model_class:
        pytest.fail(f"Model '{model_name}' not found in SCHEMA_MAP.")

    stmt = select(model_class).filter_by(**lookup_criteria)
    
    if relationships_to_load:
        for rel_name in relationships_to_load:
            if hasattr(model_class, rel_name):
                stmt = stmt.options(selectinload(getattr(model_class, rel_name)))
            else:
                pytest.fail(f"Relation '{rel_name}' not found on model '{model_name}'.")

    result = await session.execute(stmt)
    instance = result.scalar_one_or_none()
    assert instance is not None, f"Could not find instance of {model_name} with criteria {lookup_criteria}"
    return instance


def get_nested_attribute(instance, attribute_path: str):
    current_value = instance
    for part in attribute_path.split('.'):
        current_value = getattr(current_value, part)
    return current_value


@pytest.mark.asyncio
@pytest.mark.integration
@pytest.mark.unit
@pytest.mark.e2e
async def test_assertion_runner(prepared_session_factory, test_case_data):
    session_factory = prepared_session_factory
    
    for assertion_data in test_case_data["assertions"]:
        async with session_factory() as session:
            assertion_type = assertion_data["type"]
            params = assertion_data["params"]
            desc = assertion_data.get('description', assertion_type)

            logger.debug(f"--> Executing: '{desc}'")
        
            if assertion_type == "row_count":
                model_class = SCHEMA_MAP.get(params["model"])
                stmt = select(func.count()).select_from(model_class)
                if "filter_by" in params:
                    stmt = stmt.filter_by(**params["filter_by"]) 
                actual = (await session.execute(stmt)).scalar_one()
                assert actual == params["expected"], f"Expected {params['expected']} rows for {params['model']}, got {actual}"

            elif assertion_type == "attribute_equals":
                rels_to_load = [params["attribute"].split('.')[0]] if '.' in params["attribute"] else None
                instance = await get_object(session, params["model"], params["lookup"], rels_to_load)
                actual = get_nested_attribute(instance, params["attribute"])
                assert actual == params["expected"], f"Expected {params['attribute']} to be {params['expected']}, got {actual}"
            
            elif assertion_type == "collection_has_length":
                rels_to_load = [params["attribute"]]
                instance = await get_object(session, params["model"], params["lookup"], rels_to_load)
                actual_collection = get_nested_attribute(instance, params["attribute"])
                assert len(actual_collection) == params["expected"], f"Expected collection length {params['expected']}, got {len(actual_collection)}"
                
            elif assertion_type == "is_not_none":
                instance = await get_object(session, params["model"], params["lookup"])
                assert instance is not None, f"Expected instance of {params['model']} to exist"

            elif assertion_type == "execute_function":
                func_name = params["function_name"]
                func_params = params.get("function_params", {})
                
                if "on_instance" in func_params:
                    instance_config = func_params.pop("on_instance")
                    
                    instance = await get_object(
                        session,
                        instance_config["model"],
                        instance_config["lookup"],
                        instance_config.get("selectinload_relations")
                    )
                    
                    method_to_call = getattr(instance, func_name)
                    method_signature = inspect.signature(method_to_call)
                    allowed_params_names = method_signature.parameters.keys()

                    call_params = {}
                    if 'async_session' in allowed_params_names:
                        call_params['async_session'] = session_factory
                    
                    method_params = func_params.get("method_params", {})
                    for name, value in method_params.items():
                        if name not in allowed_params_names:
                            continue
                        if isinstance(value, dict) and "create_dummy" in value:
                            dummy_config = value["create_dummy"]
                            dummy_model = SCHEMA_MAP[dummy_config["model"]]
                            dummy_data = dummy_config["data"].copy()
                            parent_model_name = instance_config["model"].lower()
                            foreign_key_name = f"{parent_model_name}_id"
                            if foreign_key_name not in dummy_data:
                                dummy_data[foreign_key_name] = instance.id
                            call_params[name] = dummy_model(**dummy_data)
                        else:
                            call_params[name] = value
                    
                    result = await method_to_call(**call_params)
                
                else: 
                    target_function = FUNCTION_MAP.get(func_name)
                    if not target_function:
                        pytest.fail(f"Function '{func_name}' not found in FUNCTION_MAP.")
                    
                    call_params = func_params.copy()
                    if 'async_session' in inspect.signature(target_function).parameters:
                        call_params["async_session"] = session_factory
                    
                    result = await target_function(**call_params)
                
                if "expected_result" in params:
                    expected = params["expected_result"]
                    op = expected["operator"]
                    if op == "is_not_none":
                        assert result is not None, f"Expected result to not be None, but it was."
                    elif op == "is_none":
                        assert result is None, f"Expected result to be None, but it was not."
                    elif op == "equals":
                        assert result == expected["value"], f"Expected result '{expected['value']}', but got '{result}'."
                    else:
                        pytest.fail(f"Unknown result operator: '{op}'")

            else:
                pytest.fail(f"Unknown assertion type: '{assertion_type}'")