package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"users-service/app/models/ent"
	"users-service/tests/registry"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Step struct {
	Type          string          `json:"type"`
	Description   string          `json:"description"`
	Params        json.RawMessage `json:"params"`
	Expected      json.RawMessage `json:"expected"`
	StoreResultAs string          `json:"store_result_as"`
}

type AssertParams struct {
	Actual   json.RawMessage `json:"actual"`
	Expected json.RawMessage `json:"expected"`
	Operator string          `json:"operator"`
	Message  string          `json:"message"`
}

type CreateModelParams struct {
	ModelType string          `json:"model_type"`
	Data      json.RawMessage `json:"data"`
}

func handleFunctionExecution(t *testing.T, step Step, dependencies map[string]interface{}) {
	var rawParams struct {
		FunctionName string        `json:"function_name"`
		Args         []interface{} `json:"args"`
	}

	err := json.Unmarshal(step.Params, &rawParams)
	require.NoError(t, err, "Failed to unmarshal raw parameters for function execution")

	var funcValue reflect.Value
	var funcType reflect.Type

	targetFunc, ok := (*registry.FunctionRegistry)[rawParams.FunctionName]
	if ok {
		funcValue = reflect.ValueOf(targetFunc)
		funcType = funcValue.Type()
	} else {
		parts := strings.Split(rawParams.FunctionName, ".")
		if len(parts) == 2 {
			objectName := parts[0]
			methodName := parts[1]

			objectVal, objOk := dependencies[objectName]
			require.True(t, objOk, "Object '%s' not found in dependencies for method call '%s'", objectName, rawParams.FunctionName)

			receiver := reflect.ValueOf(objectVal)
			method := receiver.MethodByName(methodName)

			if !method.IsValid() && receiver.Kind() == reflect.Struct && receiver.CanAddr() {
				method = receiver.Addr().MethodByName(methodName)
			}

			if !method.IsValid() && receiver.Kind() == reflect.Ptr && !receiver.IsNil() {
				method = receiver.Elem().MethodByName(methodName)
			}

			if !method.IsValid() {
				t.Logf("DEBUG: Method '%s' not found. Final receiver type: %T (Kind: %s). IsValid: %t. CanAddr: %t.",
					methodName, receiver.Interface(), receiver.Kind(), method.IsValid(), receiver.CanAddr())
				if receiver.Kind() == reflect.Ptr && !receiver.IsNil() {
					t.Logf("DEBUG: Elem Type: %s, Elem Kind: %s", receiver.Elem().Type(), receiver.Elem().Kind())
				}
			}

			require.True(t, method.IsValid(), "Method '%s' not found on object '%s' (type %T). Check method name and export status, and receiver type (pointer vs value).", methodName, objectName, objectVal)

			funcValue = method
			funcType = method.Type()
		} else {
			t.Fatalf("Function/method '%s' not found in registry and does not match 'object.Method' format.", rawParams.FunctionName)
		}
	}

	require.Equal(t, len(rawParams.Args), funcType.NumIn(), "Incorrect number of arguments for function/method %s. Expected %d, Got %d.", rawParams.FunctionName, funcType.NumIn(), len(rawParams.Args))

	in := make([]reflect.Value, len(rawParams.Args))
	for i, rawArg := range rawParams.Args {
		paramType := funcType.In(i)

		if paramType == reflect.TypeOf(&gin.Context{}) {
			if rawArgString, isString := rawArg.(string); isString {
				rePlaceholder := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.-]+)\s*}}`)
				match := rePlaceholder.FindStringSubmatch(rawArgString)

				if len(match) > 1 {
					keyPath := strings.Trim(match[0], "{} ")
					directValue, directOk := dependencies[keyPath]
					require.True(t, directOk, "Dependency '%s' for function/method %s arg %d not found in dependencies.", keyPath, rawParams.FunctionName, i)
					directValReflect := reflect.ValueOf(directValue)
					require.True(t, directValReflect.Type().ConvertibleTo(paramType), "Dependency '%s' for function/method %s arg %d expects %s, but got %T.", keyPath, rawParams.FunctionName, i, paramType, directValue)

					in[i] = directValReflect.Convert(paramType)
					continue
				}
			}
			t.Fatalf("Argument %d for function/method %s expects *gin.Context, but input '%v' could not be resolved or converted.", i, rawParams.FunctionName, rawArg)
		}

		marshaledRawArg, marshalErr := json.Marshal(rawArg)
		require.NoError(t, marshalErr, "Failed to marshal argument %d for function/method %s. Value: %v", i, rawParams.FunctionName, rawArg)

		var intermediateArg interface{}
		err = json.Unmarshal(marshaledRawArg, &intermediateArg)
		require.NoError(t, err, "Failed to unmarshal intermediate argument %d for function/method %s. Bytes: %s", i, rawParams.FunctionName, string(marshaledRawArg))

		convertedArg, convErr := convertArg(reflect.ValueOf(intermediateArg), paramType)
		require.NoError(t, convErr, "Argument conversion failed for function/method %s, arg %d. Attempting to convert %T to %s. Value: %v", rawParams.FunctionName, i, intermediateArg, paramType.String(), intermediateArg)
		in[i] = convertedArg
	}

	results := funcValue.Call(in)

	if len(step.Expected) > 0 {
		var expectedReturn struct {
			ReturnValue json.RawMessage `json:"return_value"`
		}
		err = json.Unmarshal(step.Expected, &expectedReturn)
		require.NoError(t, err, "Failed to unmarshal expected return value for function execution")

		if len(results) == 0 {
			require.True(t, expectedReturn.ReturnValue == nil || string(expectedReturn.ReturnValue) == "null", "Function/method %s returned no values, but expected a non-null return value: %s", rawParams.FunctionName, string(expectedReturn.ReturnValue))
			return
		}

		actualResult := results[0].Interface()

		if err, isError := actualResult.(error); isError {
			var expectedString string
			errUnmarshal := json.Unmarshal(expectedReturn.ReturnValue, &expectedString)
			require.NoError(t, errUnmarshal, "Failed to unmarshal expected return value as string")
			assert.Equal(t, expectedString, err.Error(), "Function return value mismatch (error case)")
		} else {
			var expectedValueToCompare interface{}
			err = json.Unmarshal(expectedReturn.ReturnValue, &expectedValueToCompare)
			require.NoError(t, err, "Failed to unmarshal expected return value from RawMessage")

			actualResultType := reflect.TypeOf(actualResult)
			expectedValueToCompareReflect := reflect.ValueOf(expectedValueToCompare)

			if actualResultType != nil && expectedValueToCompareReflect.IsValid() && actualResultType != expectedValueToCompareReflect.Type() {
				if expectedValueToCompareReflect.Type().ConvertibleTo(actualResultType) {
					expectedValueToCompare = expectedValueToCompareReflect.Convert(actualResultType).Interface()
				} else if actualResultType.Kind() == reflect.Slice && expectedValueToCompareReflect.Kind() == reflect.Slice {
					marshaledExpected, marshalErr := json.Marshal(expectedValueToCompare)
					if marshalErr == nil {
						ptrToNewSlice := reflect.New(actualResultType)
						unmarshalErr := json.Unmarshal(marshaledExpected, ptrToNewSlice.Interface())
						if unmarshalErr == nil {
							expectedValueToCompare = ptrToNewSlice.Elem().Interface()
						} else {
							t.Logf("Warning: Failed to unmarshal expected slice JSON '%s' into target slice type %s: %v", string(marshaledExpected), actualResultType, unmarshalErr)
						}
					} else {
						t.Logf("Warning: Failed to marshal expected slice value %v to JSON for conversion: %v", expectedValueToCompare, marshalErr)
					}
				} else if actualResultType.Kind() == reflect.Struct && expectedValueToCompareReflect.Kind() == reflect.Map {
					marshaledExpected, marshalErr := json.Marshal(expectedValueToCompare)
					if marshalErr == nil {
						newStructPtr := reflect.New(actualResultType)
						unmarshalErr := json.Unmarshal(marshaledExpected, newStructPtr.Interface())
						if unmarshalErr == nil {
							expectedValueToCompare = newStructPtr.Elem().Interface()
						} else {
							t.Logf("Warning: Failed to unmarshal expected map JSON '%s' into target struct type %s: %v", string(marshaledExpected), actualResultType, unmarshalErr)
						}
					} else {
						t.Logf("Warning: Failed to marshal expected struct value %v to JSON for conversion: %v", expectedValueToCompare, marshalErr)
					}
				}
			}

			assert.Equal(t, expectedValueToCompare, actualResult, "Function return value mismatch")
		}
	}

	if step.StoreResultAs != "" {
		require.NotEmpty(t, results, "Function %s returned no values, cannot store result", rawParams.FunctionName)
		dependencies[step.StoreResultAs] = results[0].Interface()
	}
}

func handleCreateModel(t *testing.T, step Step, dependencies map[string]interface{}) {
	var params CreateModelParams

	err := json.Unmarshal(step.Params, &params)
	require.NoError(t, err, "Failed to unmarshal createModel step parameters")

	targetType, ok := registry.ModelTypeRegistry[params.ModelType]
	t.Logf("Attempting to create model: %s. Found in registry: %t", params.ModelType, ok)
	require.True(t, ok, "Model type '%s' not found in ModelTypeRegistry. Did you register it?", params.ModelType)

	newModelPtr := reflect.New(targetType.Elem())

	resolvedData := resolveDependencies(t, params.Data, dependencies)

	err = json.Unmarshal(resolvedData, newModelPtr.Interface())
	require.NoError(t, err, "Failed to unmarshal data into model of type '%s'", params.ModelType)

	if step.StoreResultAs != "" {
		dependencies[step.StoreResultAs] = newModelPtr.Interface()
	}
}

func handleContainsAssertion(t *testing.T, actual, expected interface{}, msg string, shouldContain bool) {
	actualVal := reflect.ValueOf(actual)
	found := false

	switch actualVal.Kind() {
	case reflect.String:
		if expectedStr, ok := expected.(string); ok {
			found = strings.Contains(actualVal.String(), expectedStr)
		} else {
			t.Fatalf("Contains operator on string requires expected value to be string. Actual type: %T, Expected type: %T", actual, expected)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < actualVal.Len(); i++ {
			if reflect.DeepEqual(actualVal.Index(i).Interface(), expected) {
				found = true
				break
			}
		}
	case reflect.Map:
		if actualMap, ok := actual.(map[string]interface{}); ok {
			if expectedKey, keyOk := expected.(string); keyOk {
				_, found = actualMap[expectedKey]
			} else {
				t.Logf("Warning: Contains operator on map currently only checks for key existence with string expected values. Expected type: %T", expected)
			}
		} else {
			t.Fatalf("Contains operator on map requires actual to be a map[string]interface{}. Actual type: %T", actual)
		}
	default:
		t.Fatalf("Contains operator is only supported for string, slice, array, or map types. Actual type: %T", actual)
	}

	if shouldContain {
		require.True(t, found, "Expected '%v' to contain '%v'. %s", actual, expected, msg)
	} else {
		require.False(t, found, "Expected '%v' not to contain '%v'. %s", actual, expected, msg)
	}
}

func handleHTTPRequest(t *testing.T, router *gin.Engine, step Step, dependencies map[string]interface{}) {
	var params struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Body   json.RawMessage `json:"body"`
	}

	resolvedParams := resolveDependencies(t, step.Params, dependencies)
	err := json.Unmarshal(resolvedParams, &params)
	require.NoError(t, err)

	req, err := http.NewRequest(params.Method, params.Path, bytes.NewBuffer(params.Body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var expected struct {
		StatusCode int             `json:"statusCode"`
		Body       json.RawMessage `json:"body"`
	}
	err = json.Unmarshal(step.Expected, &expected)
	require.NoError(t, err)

	assert.Equal(t, expected.StatusCode, w.Code, "Status code mismatch")

	if len(expected.Body) > 0 && string(expected.Body) != "null" {
		var expectedBody, actualBody interface{}
		err = json.Unmarshal(expected.Body, &expectedBody)
		require.NoError(t, err, "Failed to unmarshal expected body from JSON")

		err = json.Unmarshal(w.Body.Bytes(), &actualBody)
		require.NoError(t, err, "Failed to unmarshal actual response body")

		if expectedMap, ok := expectedBody.(map[string]interface{}); ok {
			actualMap, ok := actualBody.(map[string]interface{})
			require.True(t, ok, "Expected actual body to be a JSON object")
			require.Subset(t, actualMap, expectedMap, "Response body does not contain expected subset")
		} else {
			assert.Equal(t, expectedBody, actualBody, "Response body mismatch")
		}
	}

	if step.StoreResultAs != "" {
		var resultBody interface{}
		err = json.Unmarshal(w.Body.Bytes(), &resultBody)
		require.NoError(t, err, "Failed to unmarshal response body for storing")
		dependencies[step.StoreResultAs] = resultBody
	}
}

func handleDBRowCount(t *testing.T, tx *ent.Tx, step Step, dependencies map[string]interface{}) {
	var params struct {
		Model    string                 `json:"model"`
		FilterBy map[string]interface{} `json:"filter_by"`
		Expected int                    `json:"expected"`
	}
	resolvedParams := resolveDependencies(t, step.Params, dependencies)
	err := json.Unmarshal(resolvedParams, &params)
	require.NoError(t, err)

	manager, ok := registry.SchemaMap[params.Model]
	require.True(t, ok, "ModelManager for %s not found", params.Model)

	modelClient := manager.GetClient(tx.Client())
	query := reflect.ValueOf(modelClient).MethodByName("Query").Call(nil)[0]

	if len(params.FilterBy) > 0 {
		// TO DO
	}

	countMethod := query.MethodByName("Count")
	results := countMethod.Call([]reflect.Value{reflect.ValueOf(context.Background())})

	count := int(results[0].Int())
	err, _ = results[1].Interface().(error)
	require.NoError(t, err)

	assert.Equal(t, params.Expected, count, "Row count mismatch for model %s", params.Model)
}

func handleDBAttributeEquals(t *testing.T, tx *ent.Tx, step Step, dependencies map[string]interface{}) {
	t.Skip("dbAttributeEquals step not fully implemented yet")
}

func handleAssertion(t *testing.T, step Step, dependencies map[string]interface{}) {
	var params AssertParams

	resolvedParamsJsonBytes := resolveDependencies(t, step.Params, dependencies)
	err := json.Unmarshal(resolvedParamsJsonBytes, &params)
	require.NoError(t, err, "Failed to unmarshal assert step parameters")

	var actualValFinal, expectedValFinal interface{}

	err = json.Unmarshal(params.Actual, &actualValFinal)
	require.NoError(t, err, "Failed to unmarshal actual value for assertion. Raw: %s", string(params.Actual))

	if len(params.Expected) > 0 {
		err = json.Unmarshal(params.Expected, &expectedValFinal)
		require.NoError(t, err, "Failed to unmarshal expected value for assertion. Raw: %s", string(params.Expected))
	}

	convertToFloat64 := func(val interface{}, operator string) (float64, bool) {
		switch v := val.(type) {
		case string:
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				t.Logf("Warning: Failed to convert string '%s' to float64 for %s operator: %v", v, operator, err)
				return 0, false
			}
			return num, true
		case float64:
			return v, true
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		default:
			t.Logf("Warning: Value of type %T cannot be converted to float64 for %s operator", val, operator)
			return 0, false
		}
	}

	switch params.Operator {
	case "equals":
		require.Equal(t, expectedValFinal, actualValFinal, params.Message)
	case "notEquals":
		require.NotEqual(t, expectedValFinal, actualValFinal, params.Message)
	case "length":
		actualReflectVal := reflect.ValueOf(actualValFinal)
		if actualReflectVal.Kind() == reflect.Invalid {
			t.Fatalf("Length operator requires a valid type, but actual value is invalid/nil")
		}
		expectedLenFloat, expectedLenOk := convertToFloat64(expectedValFinal, "length")
		require.True(t, expectedLenOk, "Length operator requires numeric expected value, but got %T (value: %v)", expectedValFinal, expectedValFinal)
		require.Len(t, actualValFinal, int(expectedLenFloat), params.Message)
	case "greaterThan":
		actualFloat, actualOk := convertToFloat64(actualValFinal, "greaterThan")
		expectedFloat, expectedOk := convertToFloat64(expectedValFinal, "greaterThan")
		require.True(t, actualOk && expectedOk, "Assertion operator '%s' requires numeric types, but got actual=%T (value: %v), expected=%T (value: %v)",
			params.Operator, actualValFinal, actualValFinal, expectedValFinal, expectedValFinal)
		require.Greater(t, actualFloat, expectedFloat, params.Message)
	case "lessThan":
		actualFloat, actualOk := convertToFloat64(actualValFinal, "lessThan")
		expectedFloat, expectedOk := convertToFloat64(expectedValFinal, "lessThan")
		require.True(t, actualOk && expectedOk, "Assertion operator '%s' requires numeric types, but got actual=%T (value: %v), expected=%T (value: %v)",
			params.Operator, actualValFinal, actualValFinal, expectedValFinal, expectedValFinal)
		require.Less(t, actualFloat, expectedFloat, params.Message)
	case "contains":
		handleContainsAssertion(t, actualValFinal, expectedValFinal, params.Message, true)
	case "notContains":
		handleContainsAssertion(t, actualValFinal, expectedValFinal, params.Message, false)
	case "isNil":
		require.Nil(t, actualValFinal, params.Message)
	case "notNil":
		require.NotNil(t, actualValFinal, params.Message)
	case "isEmpty":
		require.Empty(t, actualValFinal, params.Message)
	case "isNotEmpty":
		require.NotEmpty(t, actualValFinal, params.Message)
	default:
		t.Fatalf("Unknown assertion operator: %s", params.Operator)
	}
}

func convertArg(argValue reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	if argValue.Type().ConvertibleTo(targetType) {
		return argValue.Convert(targetType), nil
	}

	if argValue.Kind() == reflect.Float64 && (targetType.Kind() == reflect.Int || targetType.Kind() == reflect.Int64) {
		return reflect.ValueOf(int64(argValue.Float())).Convert(targetType), nil
	}

	if argValue.Kind() == reflect.Slice && targetType.Kind() == reflect.Slice {
		targetSliceType := targetType.Elem()
		if argValue.IsNil() {
			return reflect.Zero(targetType), nil
		}
		sourceSlice := argValue.Interface().([]interface{})

		newSlice := reflect.MakeSlice(targetType, len(sourceSlice), len(sourceSlice))

		for i, v := range sourceSlice {
			elemValue := reflect.ValueOf(v)

			if elemValue.Kind() == reflect.Map {

				tempJson, err := json.Marshal(v)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to re-marshal map to json: %w", err)
				}

				newElemPtr := reflect.New(targetSliceType)
				if targetSliceType.Kind() == reflect.Ptr && targetSliceType.Elem().Kind() == reflect.Struct {
					newElemPtr = reflect.New(targetSliceType.Elem())
				}

				err = json.Unmarshal(tempJson, newElemPtr.Interface())
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to unmarshal json to target struct %s: %w", targetSliceType, err)
				}

				if targetSliceType.Kind() == reflect.Ptr && targetSliceType.Elem().Kind() == reflect.Struct {
					newSlice.Index(i).Set(newElemPtr)
				} else {
					newSlice.Index(i).Set(newElemPtr.Elem())
				}
			} else if elemValue.Type().ConvertibleTo(targetSliceType) {
				newSlice.Index(i).Set(elemValue.Convert(targetSliceType))
			} else {
				return reflect.Value{}, fmt.Errorf("cannot convert slice element from %T to %s", v, targetSliceType)
			}
		}
		return newSlice, nil
	}

	if argValue.Kind() == reflect.Map && targetType.Kind() == reflect.Ptr && targetType.Elem().Kind() == reflect.Struct {
		jsonData, err := json.Marshal(argValue.Interface())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to marshal map[string]interface{} to JSON: %w", err)
		}
		newStructPtr := reflect.New(targetType.Elem())
		err = json.Unmarshal(jsonData, newStructPtr.Interface())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to unmarshal JSON into %s: %w", targetType, err)
		}
		return newStructPtr, nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert argument from %s to %s", argValue.Type(), targetType)
}

func resolveSingleDependencyGoValue(t *testing.T, jsonStringOrPlaceholder string, dependencies map[string]interface{}) interface{} {
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.-]+)\s*}}`)
	match := re.FindStringSubmatch(jsonStringOrPlaceholder)

	if len(match) > 1 {
		keyPath := strings.Trim(match[0], "{} ")
		keys := strings.Split(keyPath, ".")

		value, ok := dependencies[keys[0]]
		require.True(t, ok, "Dependency '%s' not found for placeholder '%s'", keys[0], jsonStringOrPlaceholder)

		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		currentVal := v.Interface()

		for i := 1; i < len(keys); i++ {
			v = reflect.ValueOf(currentVal)
			if v.Kind() == reflect.Struct {
				field := v.FieldByName(keys[i])
				require.True(t, field.IsValid(), "Field '%s' not found in struct %T for dependency '%s'", keys[i], currentVal, strings.Join(keys[:i+1], "."))
				if field.Kind() == reflect.Ptr && !field.IsNil() {
					currentVal = field.Elem().Interface()
				} else {
					currentVal = field.Interface()
				}
			} else if v.Kind() == reflect.Map {
				nestedValue := v.MapIndex(reflect.ValueOf(keys[i]))
				require.True(t, nestedValue.IsValid(), "Key '%s' not found in map %T for dependency '%s'", keys[i], currentVal, strings.Join(keys[:i+1], "."))
				currentVal = nestedValue.Interface()
			} else if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && (regexp.MustCompile(`^\d+$`).MatchString(keys[i])) {
				idx, convErr := strconv.Atoi(keys[i])
				require.NoError(t, convErr, "Invalid array index '%s' in path for dependency '%s'", keys[i], strings.Join(keys[:i+1], "."))
				require.True(t, idx >= 0 && idx < v.Len(), "Array index out of bounds: %d for array of length %d in dependency '%s'", idx, v.Len(), strings.Join(keys[:i+1], "."))
				elem := v.Index(idx)
				if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					currentVal = elem.Elem().Interface()
				} else {
					currentVal = elem.Interface()
				}
			} else {
				t.Fatalf("Cannot access '%s'; dependency '%s' is not a struct, map, or indexable slice/array (it's %T).", keys[i], strings.Join(keys[:i], "."), currentVal)
			}
		}
		return currentVal
	} else {
		var literalValue interface{}
		err := json.Unmarshal([]byte(jsonStringOrPlaceholder), &literalValue)
		require.NoError(t, err, "Failed to unmarshal direct JSON literal '%s'", jsonStringOrPlaceholder)
		return literalValue
	}
}

func resolveDependencies(t *testing.T, rawData json.RawMessage, dependencies map[string]interface{}) json.RawMessage {
	dataStr := string(rawData)
	re := regexp.MustCompile(`"{{\s*([a-zA-Z0-9_.-]+)\s*}}"`)
	resolvedFinalJsonStr := re.ReplaceAllStringFunc(dataStr, func(match string) string {
		keyPath := strings.Trim(match, "{} \"")
		keys := strings.Split(keyPath, ".")
		value, ok := dependencies[keys[0]]
		t.Logf("Resolving dependency '%s': found=%t, value=%v", keyPath, ok, value)
		require.True(t, ok, "Dependency '%s' not found for placeholder '%s'", keys[0], match)

		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		currentVal := v.Interface()

		for i := 1; i < len(keys); i++ {
			v = reflect.ValueOf(currentVal)
			if v.Kind() == reflect.Struct {
				field := v.FieldByName(keys[i])
				require.True(t, field.IsValid(), "Field '%s' not found in struct %T for dependency '%s'", keys[i], currentVal, strings.Join(keys[:i+1], "."))
				if field.Kind() == reflect.Ptr && !field.IsNil() {
					currentVal = field.Elem().Interface()
				} else {
					currentVal = field.Interface()
				}

			} else if v.Kind() == reflect.Map {
				nestedValue := v.MapIndex(reflect.ValueOf(keys[i]))
				require.True(t, nestedValue.IsValid(), "Key '%s' not found in map %T for dependency '%s'", keys[i], currentVal, strings.Join(keys[:i+1], "."))
				currentVal = nestedValue.Interface()
			} else if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && (regexp.MustCompile(`^\d+$`).MatchString(keys[i])) {
				idx, convErr := strconv.Atoi(keys[i])
				require.NoError(t, convErr, "Invalid array index '%s' in path for dependency '%s'", keys[i], strings.Join(keys[:i+1], "."))
				require.True(t, idx >= 0 && idx < v.Len(), "Array index out of bounds: %d for array of length %d in dependency '%s'", idx, v.Len(), strings.Join(keys[:i+1], "."))
				elem := v.Index(idx)
				if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					currentVal = elem.Elem().Interface()
				} else {
					currentVal = elem.Interface()
				}
			} else {
				t.Fatalf("Cannot access '%s'; dependency '%s' is not a struct, map, or indexable slice/array (it's %T).", keys[i], strings.Join(keys[:i], "."), currentVal)
			}
		}

		marshaled, err := json.Marshal(currentVal)
		require.NoError(t, err)
		return string(marshaled)
	})

	t.Logf("Final resolved JSON: %s", resolvedFinalJsonStr)
	return json.RawMessage(resolvedFinalJsonStr)
}

func parseJsonStringContentToNativeType(t *testing.T, jsonContentString string) (interface{}, error) {
	var result interface{}

	if err := json.Unmarshal([]byte(jsonContentString), &result); err == nil {
		return result, nil
	}

	return jsonContentString, nil
}
