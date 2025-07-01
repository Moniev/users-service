package registry

import "users-service/app/utils"

type FunctionContainer map[string]interface{}

var functionRegistry = &FunctionContainer{
	"utils.CheckEmailFormat":    utils.CheckEmailFormat,
	"utils.CheckPasswordFormat": utils.CheckPasswordFormat,
	"utils.CheckTokenFormat":    utils.CheckTokenFormat,
	"utils.CheckPhoneFormat":    utils.CheckPhoneFormat,
	"utils.CheckIBANFormat":     utils.CheckIBANFormat,
	"utils.CheckSWIFTFormat":    utils.CheckSWIFTFormat,
	"utils.GenerateRandomCode":  utils.GenerateRandomCode,
	"utils.IsSuccessStatusCode": utils.IsSuccessStatusCode,
}
