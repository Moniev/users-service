package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"users-service/app/models/ent"
	"users-service/tests/registry"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	SetupData []map[string]interface{} `json:"setup_data"`
	TestCases []TestCase               `json:"test_cases"`
}

type TestCase struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Type          string          `json:"type"`
	Description   string          `json:"description"`
	Params        json.RawMessage `json:"params"`
	Expected      json.RawMessage `json:"expected"`
	StoreResultAs string          `json:"store_result_as"`
}

func TestDataDriven(t *testing.T) {
	var router *gin.Engine
	if TestApp != nil {
		router = TestApp.Router
	} else {
		gin.SetMode(gin.TestMode)
		router = gin.New()
	}

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")
	basePath := filepath.Dir(currentFile)
	testDir := filepath.Join(basePath, "resources", *TestType)

	t.Logf("Scanning for test files in absolute path: %s", testDir)

	var testFiles []string
	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			testFiles = append(testFiles, path)
		}
		return nil
	})

	require.NoError(t, err, "Failed to walk test case directory: %s", testDir)
	if len(testFiles) == 0 {
		t.Skipf("Skipping: No test files found in directory: %s", testDir)
		return
	}
	t.Logf("Found %d test file(s) to execute.", len(testFiles))

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			suiteContent, err := ioutil.ReadFile(file)
			require.NoError(t, err)

			var suite TestSuite
			err = json.Unmarshal(suiteContent, &suite)
			require.NoError(t, err)

			for _, tc := range suite.TestCases {
				t.Run(tc.Name, func(t *testing.T) {
					var tx *ent.Tx
					if *TestType == "integration" || *TestType == "e2e" {
						require.NotNil(t, TestDB, "Database connection (TestDB) is nil. Ensure tests are run with integration or e2e build tags.")
						tx, err = TestDB.Tx(context.Background())
						require.NoError(t, err)
						defer tx.Rollback()
						registry.SetupDatabase(t, tx, suite.SetupData)
					}

					dependencies := make(map[string]interface{})

					for i, step := range tc.Steps {
						t.Run(fmt.Sprintf("Step %d: %s", i+1, step.Description), func(t *testing.T) {
							executeStep(t, router, tx, step, dependencies)
						})
					}
				})
			}
		})
	}
}

func executeStep(t *testing.T, router *gin.Engine, tx *ent.Tx, step Step, dependencies map[string]interface{}) {
	switch step.Type {
	case "httpRequest":
		handleHTTPRequest(t, router, step, dependencies)
	case "dbRowCount":
		require.NotNil(t, tx, "Database transaction is required for dbRowCount step, but it's nil. Use this step only for integration/e2e tests.")
		handleDBRowCount(t, tx, step, dependencies)
	case "dbAttributeEquals":
		require.NotNil(t, tx, "Database transaction is required for dbAttributeEquals step, but it's nil. Use this step only for integration/e2e tests.")
		handleDBAttributeEquals(t, tx, step, dependencies)
	case "executeFunction":
		handleFunctionExecution(t, step, dependencies)
	default:
		t.Fatalf("Unknown step type: %s", step.Type)
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
			assert.Subset(t, actualMap, expectedMap, "Response body does not contain expected subset")
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

func resolveDependencies(t *testing.T, rawData json.RawMessage, dependencies map[string]interface{}) []byte {
	dataStr := string(rawData)
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.-]+)\s*}}`)

	resolvedStr := re.ReplaceAllStringFunc(dataStr, func(match string) string {
		keyPath := strings.Trim(match, "{} ")
		keys := strings.Split(keyPath, ".")

		value, ok := dependencies[keys[0]]
		if !ok {
			t.Fatalf("Dependency '%s' not found for placeholder '%s'", keys[0], match)
		}

		for i := 1; i < len(keys); i++ {
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Map {
				nestedValue := v.MapIndex(reflect.ValueOf(keys[i]))
				if !nestedValue.IsValid() {
					t.Fatalf("Key '%s' not found in dependency '%s'", keys[i], strings.Join(keys[:i], "."))
				}
				value = nestedValue.Interface()
			} else {
				t.Fatalf("Cannot access field '%s'; dependency '%s' is not a map.", keys[i], strings.Join(keys[:i], "."))
			}
		}

		marshaled, err := json.Marshal(value)
		require.NoError(t, err)
		if strings.HasPrefix(string(marshaled), `"`) {
			return string(marshaled[1 : len(marshaled)-1])
		}

		return string(marshaled)
	})

	return []byte(resolvedStr)
}

func handleFunctionExecution(t *testing.T, step Step, dependencies map[string]interface{}) {
	var params struct {
		FunctionName string        `json:"function_name"`
		Args         []interface{} `json:"args"`
	}
	resolvedParams := resolveDependencies(t, step.Params, dependencies)
	err := json.Unmarshal(resolvedParams, &params)
	require.NoError(t, err)

	targetFunc, ok := (*registry.FunctionRegistry)[params.FunctionName]
	require.True(t, ok, "Function %s not found in FunctionRegistry", params.FunctionName)

	funcValue := reflect.ValueOf(targetFunc)
	funcType := funcValue.Type()

	require.Equal(t, len(params.Args), funcType.NumIn(), "Incorrect number of arguments for function %s", params.FunctionName)

	in := make([]reflect.Value, len(params.Args))
	for i, arg := range params.Args {
		argValue := reflect.ValueOf(arg)
		paramType := funcType.In(i)

		if argValue.Kind() == reflect.Float64 && (paramType.Kind() == reflect.Int || paramType.Kind() == reflect.Int64) {
			argValue = reflect.ValueOf(int64(argValue.Float())).Convert(paramType)
		} else if !argValue.Type().ConvertibleTo(paramType) {
			t.Fatalf("Cannot convert argument %d for function %s from %T to %s", i, params.FunctionName, arg, paramType)
		}
		in[i] = argValue.Convert(paramType)
	}

	results := funcValue.Call(in)

	var expected struct {
		ReturnValue interface{} `json:"return_value"`
	}
	if len(step.Expected) > 0 {
		err = json.Unmarshal(step.Expected, &expected)
		require.NoError(t, err)
		require.NotEmpty(t, results, "Function %s returned no values but expected one", params.FunctionName)
		actualResult := results[0].Interface()
		assert.Equal(t, expected.ReturnValue, actualResult, "Function return value mismatch")
	}

	if step.StoreResultAs != "" {
		require.NotEmpty(t, results, "Function %s returned no values, cannot store result", params.FunctionName)
		dependencies[step.StoreResultAs] = results[0].Interface()
	}
}
