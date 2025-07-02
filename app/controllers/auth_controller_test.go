//go:build unit

package controllers

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
}
