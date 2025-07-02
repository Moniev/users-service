package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"users-service/app/models/ent"
	"users-service/tests/registry"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	SetupData []map[string]interface{} `json:"setup_data"`
	TestCases []TestCase               `json:"test_cases"`
}

type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty,"`
	Author      string `json:"author,omitempty,"`
	Steps       []Step `json:"steps"`
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
			if err != nil {
				t.Logf("Warning: Could not read file %s, skipping. Error: %v", file, err)
				return
			}

			if len(strings.TrimSpace(string(suiteContent))) == 0 {
				t.Logf("Warning: Test file %s is empty or contains only whitespace, skipping.", file)
				return
			}

			var suite TestSuite
			err = json.Unmarshal(suiteContent, &suite)
			if err != nil {
				t.Logf("Warning: Could not unmarshal JSON from %s, skipping. Error: %v", file, err)
				return
			}

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
	case "assert":
		handleAssertion(t, step, dependencies)
	case "createModel":
		handleCreateModel(t, step, dependencies)
	default:
		t.Fatalf("Unknown step type: %s", step.Type)
	}
}
