package registry

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"users-service/app/models/requests"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
)

type FunctionContainer *map[string]interface{}

func CreateTestContext(values map[string]interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	for key, val := range values {
		ctx.Set(key, val)
	}
	return ctx, w
}

func GinContextWithData(contextData map[string]interface{}) *gin.Context {
	ctx, _ := CreateTestContext(contextData)
	return ctx
}

var FunctionRegistry FunctionContainer

func init() {
	if FunctionRegistry == nil {
		FunctionRegistry = &map[string]interface{}{}
	}

	(*FunctionRegistry)["utils.CheckEmailFormat"] = utils.CheckEmailFormat
	(*FunctionRegistry)["utils.CheckPasswordFormat"] = utils.CheckPasswordFormat
	(*FunctionRegistry)["utils.CheckTokenFormat"] = utils.CheckTokenFormat
	(*FunctionRegistry)["utils.CheckPhoneFormat"] = utils.CheckPhoneFormat
	(*FunctionRegistry)["utils.CheckIBANFormat"] = utils.CheckIBANFormat
	(*FunctionRegistry)["utils.CheckSWIFTFormat"] = utils.CheckSWIFTFormat
	(*FunctionRegistry)["utils.IsSuccessStatusCode"] = utils.IsSuccessStatusCode
	(*FunctionRegistry)["utils.Contains"] = utils.Contains
	(*FunctionRegistry)["utils.ValidateRoles"] = utils.ValidateRoles
	(*FunctionRegistry)["utils.GenerateRandomCode"] = utils.GenerateRandomCode
	(*FunctionRegistry)["utils.GenerateRandomCodes"] = utils.GenerateRandomCodes
	(*FunctionRegistry)["utils.GetUserID"] = utils.GetUserID
	(*FunctionRegistry)["utils.BindRoles"] = utils.BindRoles
	(*FunctionRegistry)["utils.GetToken"] = utils.GetToken

	(*FunctionRegistry)["tests.GinContextWithData"] = GinContextWithData
	(*FunctionRegistry)["requests.Register.Valid"] = func(args []reflect.Value) ([]reflect.Value, error) {
		if len(args) != 1 || !args[0].CanInterface() {
			return nil, fmt.Errorf("invalid arguments for requests.Register.Valid")
		}
		req, ok := args[0].Interface().(*requests.Register)
		if !ok {
			return nil, fmt.Errorf("argument for requests.Register.Valid must be *requests.Register, got %T", args[0].Interface())
		}
		validationErr := req.Valid()
		if validationErr != nil {
			return []reflect.Value{reflect.ValueOf(validationErr.Error())}, nil
		}
		return []reflect.Value{reflect.ValueOf(nil)}, nil
	}
}
