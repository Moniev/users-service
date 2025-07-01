package models

import "github.com/gin-gonic/gin"

type AppWrapper struct {
	Router *gin.Engine
}

// CachePayload represents the structure for encrypted user cache data.
type CachePayload struct {
	CipherText string `json:"cipher_text"` // CipherText contains the encrypted data
	Nonce      string `json:"nonce"`       // Nonce contains the cryptographic nonce used for encryption
}
