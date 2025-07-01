package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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

var testType = flag.String("test_type", "", "type of test to run (unit, integration, e2e)")

func TestDataDriven(t *testing.T) {
	flag.Parse()

	if *testType == "" {
		t.Skip("Skipping data-driven tests: -test_type flag not provided")
		return
	}

	var router *gin.Engine
	if TestApp != nil {
		router = TestApp.Router
	} else {
		gin.SetMode(gin.TestMode)
		router = gin.New()
	}

	testDir := filepath.Join("resources", *testType)
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
					if TestDB != nil {
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
		handleDBRowCount(t, tx, step, dependencies)
	case "dbAttributeEquals":
		handleDBAttributeEquals(t, tx, step, dependencies)
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
