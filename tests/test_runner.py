import pytest
from loguru import logger
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
import inspect
from typing import Any, Dict

from tests.conftest import FUNCTION_MAP, SCHEMA_MAP, SERVICE_MAP, SERVICE_DEPENDENCIES


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


async def check_expected_result(result, expected_params):
    op = expected_params["operator"]
    
    if op == "is_not_none":
        assert result is not None, "Expected result to not be None, but it was."
    elif op == "is_none":
        assert result is None, "Expected result to be None, but it was not."
    elif op == "equals":
        expected_value = expected_params["value"]
        if isinstance(result, dict) and isinstance(expected_value, dict):
            for key, val in expected_value.items():
                assert key in result, f"Expected key '{key}' not in result dictionary."
                assert result[key] == val, f"For key '{key}', expected '{val}', but got '{result[key]}'."
        else:
            assert result == expected_value, f"Expected result '{expected_value}', but got '{result}'."
    else:
        pytest.fail(f"Unknown result operator: '{op}'")


def _resolve_params_from_dependencies(params: Dict[str, Any], dependencies: Dict[str, Any]) -> Dict[str, Any]:
    resolved_params = {}
    for key, value in params.items():
        if isinstance(value, dict) and "from_dependency" in value:
            dependency_key = value["from_dependency"]
            if dependency_key not in dependencies:
                pytest.fail(f"Dependency '{dependency_key}' not found for parameter '{key}'.")
            resolved_params[key] = dependencies[dependency_key]
        elif isinstance(value, dict):
            resolved_params[key] = _resolve_params_from_dependencies(value, dependencies)
        elif isinstance(value, list):
            resolved_params[key] = [
                _resolve_params_from_dependencies(item, dependencies) if isinstance(item, dict) else item
                for item in value
            ]
        else:
            resolved_params[key] = value
    return resolved_params


@pytest.mark.asyncio
@pytest.mark.integration
@pytest.mark.unit
@pytest.mark.e2e
async def test_assertion_runner(prepared_session_factory, test_case_data, kafka_producer_session):
    session_factory = prepared_session_factory
    dependency_instances = {}

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
                
                if "expected_result_key" in params:
                    dependency_instances[params["expected_result_key"]] = actual
                    logger.debug(f"Stored '{params['attribute']}' ({actual}) into dependency '{params['expected_result_key']}'.")
                elif "expected" in params:
                    assert actual == params["expected"], f"Expected {params['attribute']} to be {params['expected']}, got {actual}"
                else:
                    pytest.fail(f"AttributeEquals assertion for '{params['attribute']}' is missing 'expected' or 'expected_result_key' in params.")
            
            elif assertion_type == "collection_has_length":
                rels_to_load = [params["attribute"]]
                instance = await get_object(session, params["model"], params["lookup"], rels_to_load)
                actual_collection = get_nested_attribute(instance, params["attribute"])
                assert len(actual_collection) == params["expected"], f"Expected collection length {params['expected']}, got {len(actual_collection)}"
                
            elif assertion_type == "is_not_none":
                instance = await get_object(session, params["model"], params["lookup"])
                assert instance is not None, f"Expected instance of {params['model']} to exist"

            elif assertion_type == "execute_service_function":
                service_name = params["service"]
                func_name = params["function_name"]
                
                resolved_func_params = _resolve_params_from_dependencies(params.get("function_params", {}), dependency_instances)

                service_class = SERVICE_MAP.get(service_name)
                if not service_class:
                    pytest.fail(f"Service '{service_name}' not found in SERVICE_MAP.")

                service_kwargs = {}
                if service_class in SERVICE_DEPENDENCIES:
                    service_kwargs = SERVICE_DEPENDENCIES[service_class].copy()

                service_kwargs['session_factory'] = session_factory

                service_signature = inspect.signature(service_class)
                if 'kafka_producer' in service_signature.parameters:
                    service_kwargs['kafka_producer'] = kafka_producer_session
                
                service_instance = service_class(**service_kwargs)
                method_to_call = getattr(service_instance, func_name)
                
                result = await method_to_call(**resolved_func_params)

                if "expected_result" in params:
                    await check_expected_result(result, params["expected_result"])

            elif assertion_type == "execute_function":
                func_name = params["function_name"]
                
                # Resolve function parameters from dependencies
                resolved_func_params = _resolve_params_from_dependencies(params.get("function_params", {}), dependency_instances)

                if "on_instance" in resolved_func_params:
                    instance_config = resolved_func_params.pop("on_instance")
                    
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
                    
                    method_params = resolved_func_params.get("method_params", {}) # Use resolved_func_params
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
                    
                    call_params = resolved_func_params.copy() # Use resolved_func_params
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