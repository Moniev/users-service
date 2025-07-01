package registry

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"users-service/app/models/ent"

	"github.com/stretchr/testify/require"
)

type ModelManager interface {
	GetClient(client *ent.Client) interface{}
}

type GenericModelManager struct {
	ClientFunc func(client *ent.Client) interface{}
}

func (m *GenericModelManager) GetClient(c *ent.Client) interface{} {
	return m.ClientFunc(c)
}

var SchemaMap = map[string]ModelManager{
	"User":             &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.User }},
	"UserRole":         &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.UserRole }},
	"UserDevice":       &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.UserDevice }},
	"UserSettings":     &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.UserSettings }},
	"UserDetails":      &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.UserDetails }},
	"ActivationCode":   &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.ActivationCode }},
	"ResetCode":        &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.ResetCode }},
	"RolePermission":   &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.RolePermission }},
	"SecondFactorCode": &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.SecondFactorCode }},
	"UserAction":       &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.UserAction }},
	"VerificationCode": &GenericModelManager{ClientFunc: func(c *ent.Client) interface{} { return c.VerificationCode }},
}

func SetupDatabase(t *testing.T, client *ent.Tx, setupData []map[string]interface{}) {
	if len(setupData) == 0 {
		return
	}

	ctx := context.Background()
	createdEntities := make(map[string]map[interface{}]int)

	for _, item := range setupData {
		modelName, _ := item["model"].(string)
		data, _ := item["data"].(map[string]interface{})
		localID, hasLocalID := data["id"]

		manager, ok := SchemaMap[modelName]
		require.True(t, ok, "ModelManager for %s not found in SchemaMap", modelName)

		modelClient := manager.GetClient(client.Client())
		builder := reflect.ValueOf(modelClient).MethodByName("Create").Call(nil)[0]

		for key, val := range data {
			if key == "id" {
				continue
			}

			var methodName string
			var values []interface{}

			if strings.HasSuffix(key, "_ids") {
				singularKey := strings.TrimSuffix(key, "s")
				methodName = "Add" + toCamelCase(singularKey) + "s"

				localIDs, ok := val.([]interface{})
				require.True(t, ok, "Expected a list of local IDs for key %s", key)

				realIDs := make([]int, len(localIDs))
				dependentModelName := toCamelCase(strings.TrimSuffix(key, "_ids"))
				for i, lid := range localIDs {
					realID, found := createdEntities[dependentModelName][lid]
					require.True(t, found, "Dependency not found for %s with local ID %v", dependentModelName, lid)
					realIDs[i] = realID
				}
				values = []interface{}{realIDs}

			} else if strings.HasSuffix(key, "_id") {
				methodName = "Set" + toCamelCase(key)

				dependentModelName := toCamelCase(strings.TrimSuffix(key, "_id"))
				realID, found := createdEntities[dependentModelName][val]
				require.True(t, found, "Dependency not found for %s with local ID %v", dependentModelName, val)
				values = []interface{}{realID}

			} else {
				methodName = "Set" + toCamelCase(key)
				values = []interface{}{val}
			}

			method := builder.MethodByName(methodName)
			if !method.IsValid() {
				t.Logf("Warning: Method %s not found on builder for model %s. Skipping field.", methodName, modelName)
				continue
			}

			args := []reflect.Value{}
			for i, v := range values {
				paramType := method.Type().In(i)
				argValue := reflect.ValueOf(v)
				if argValue.Type().ConvertibleTo(paramType) {
					args = append(args, argValue.Convert(paramType))
				} else {
					t.Fatalf("Cannot convert value for %s.%s from %T to %s", modelName, key, v, paramType)
				}
			}

			builder = method.Call(args)[0]
		}

		saveMethod := builder.MethodByName("Save")
		results := saveMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})

		errResult := results[1].Interface()
		if errResult != nil {
			err, ok := errResult.(error)
			require.True(t, ok)
			require.NoError(t, err, "Failed to save entity %s", modelName)
		}

		if hasLocalID {
			savedEntity := results[0].Interface()
			newEntityID := reflect.ValueOf(savedEntity).Elem().FieldByName("ID").Int()

			if _, ok := createdEntities[modelName]; !ok {
				createdEntities[modelName] = make(map[interface{}]int)
			}
			createdEntities[modelName][localID] = int(newEntityID)
		}
	}
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func cleanupDatabase(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	_, err := client.ActivationCode.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.SecondFactorCode.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.ResetCode.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.VerificationCode.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.UserDevice.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.UserAction.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.UserSettings.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.UserDetails.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.User.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.UserRole.Delete().Exec(ctx)
	require.NoError(t, err)
	_, err = client.RolePermission.Delete().Exec(ctx)
	require.NoError(t, err)
}
