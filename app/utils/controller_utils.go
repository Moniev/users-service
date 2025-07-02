package utils

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	randGen "math/rand/v2"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// GetRequestID retrieves the RequestID from the Gin context.
func GetRequestID(ctx *gin.Context) string {
	requestIDraw, exists := ctx.Get("RequestID")
	if !exists {
		return ""
	}
	requestID, ok := requestIDraw.(string)
	if !ok {
		return ""
	}
	return requestID
}

// Contains checks if a target string exists within a slice of strings.
func Contains(slice []string, target string) bool {
	return slices.Contains(slice, target)
}

// GenerateRandomCode generates a random 6-digit integer between 100000 and 999999.
func GenerateRandomCode() int {
	return randGen.IntN(900000) + 100000
}

// GenerateRandomCodes generates a slice of random 6-digit integers.
func GenerateRandomCodes(n int) []int {
	codes := make([]int, n)
	for i := 0; i < n; i++ {
		codes[i] = GenerateRandomCode()
	}
	return codes
}

func IsSuccessStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return true
	default:
		return false
	}
}

func LoadKeys(jwtPrivateKeyPath, jwtPublicKeyPath string) (crypto.PrivateKey, crypto.PublicKey, error) {
	privFilePath := jwtPrivateKeyPath
	privatePEM, err := os.ReadFile(privFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read private key file at %s: %w", privFilePath, err)
	}

	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return nil, nil, errors.New("failed to decode PEM block containing private key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	edPrivateKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, errors.New("private key is not of type Ed25519")
	}

	pubFilePath := jwtPublicKeyPath
	publicPEM, err := os.ReadFile(pubFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read public key file at %s: %w", pubFilePath, err)
	}

	publicBlock, _ := pem.Decode(publicPEM)
	if publicBlock == nil {
		return nil, nil, errors.New("failed to decode PEM block containing public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	edPublicKey, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return nil, nil, errors.New("public key is not of type Ed25519")
	}

	return edPrivateKey, edPublicKey, nil
}

// CheckEmailFormat validates the format of an email address.
// It checks if the email meets basic RFC 5322 requirements, including length constraints
// and syntax rules. The function ensures the email is non-empty, not excessively long,
// matches a standard regex pattern, has exactly one "@" symbol, and adheres to additional
// rules like maximum local part length and no consecutive dots in the domain.
//
// Parameters:
//   - mail: The email address string to validate.
//
// Returns:
//   - bool: True if the email format is valid, false otherwise.
func CheckEmailFormat(mail string) bool {
	if len(mail) == 0 || len(mail) > 254 {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,63}$`)
	if !emailRegex.MatchString(mail) {
		return false
	}

	parts := strings.Split(mail, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) > 64 {
		return false
	}

	if strings.Contains(parts[1], "..") {
		return false
	}

	return true
}

// CheckPasswordFormat validates the format and strength of a password.
// It ensures the password meets security requirements by checking its length
// (between 12 and 128 characters), disallowing the "$" character, and requiring
// at least one special character (from a predefined set), one digit, and one
// uppercase letter.
//
// Parameters:
//   - password: The password string to validate.
//
// Returns:
//   - bool: True if the password meets all format requirements, false otherwise.
func CheckPasswordFormat(password string) bool {
	if len(password) < 12 || len(password) > 128 {
		return false
	}

	if strings.Contains(password, "$") {
		return false
	}

	specialCharRegex := regexp.MustCompile(`[!@#%^&*(),.?":{}|<>]`)
	hasSpecial := specialCharRegex.MatchString(password)
	hasDigit := false
	hasUpper := false

	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if unicode.IsUpper(char) {
			hasUpper = true
		}
	}

	return hasSpecial && hasDigit && hasUpper
}

// CheckTokenFormat validates the format of a token string.
// It ensures the token is exactly 6 characters long and contains only digits (0-9).
// This is typically used for verifying formats like one-time passwords (OTPs) or
// numeric verification codes.
//
// Parameters:
//   - token: The token string to validate.
//
// Returns:
//   - bool: True if the token is a 6-digit numeric string, false otherwise.
func CheckTokenFormat(token string) bool {
	if len(token) != 6 {
		return false
	}

	for _, char := range token {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// CheckPhoneFormat validates the format of a phone number.
// It ensures the phone number follows a common international format, starting with "+"
// followed by 9 to 15 digits (allowing for country codes and standard phone number lengths).
// The function checks for correct length, proper prefix, and numeric-only content after the prefix.
//
// Parameters:
//   - phone: The phone number string to validate.
//
// Returns:
//   - bool: True if the phone number format is valid, false otherwise.
func CheckPhoneFormat(phone string) bool {
	if len(phone) < 10 || len(phone) > 16 {
		return false
	}

	if !strings.HasPrefix(phone, "+") {
		return false
	}

	numberPart := phone[1:]
	if len(numberPart) == 0 {
		return false
	}

	for _, char := range numberPart {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// CheckIBANFormat validates the format of an IBAN (International Bank Account Number).
// It ensures the IBAN follows the ISO 13616 standard: starts with a 2-letter country code,
// followed by 2 check digits, and then an alphanumeric basic bank account number (BBAN).
// The function checks length constraints (15-34 characters), correct prefix format,
// and performs a basic structure validation using a regex pattern.
//
// Parameters:
//   - iban: The IBAN string to validate.
//
// Returns:
//   - bool: True if the IBAN format is valid, false otherwise.
func CheckIBANFormat(iban string) bool {
	iban = strings.ReplaceAll(iban, " ", "")

	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	ibanRegex := regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$`)
	return ibanRegex.MatchString(iban)
}

// CheckSWIFTFormat validates the format of a SWIFT/BIC code.
// It ensures the SWIFT/BIC follows the ISO 9362 standard: 8 or 11 characters long,
// with a specific structure: 4-letter bank code, 2-letter country code, 2-character
// location code, and an optional 3-character branch code. Only uppercase letters
// and digits are allowed.
//
// Parameters:
//   - swift: The SWIFT/BIC string to validate.
//
// Returns:
//   - bool: True if the SWIFT/BIC format is valid, false otherwise.
func CheckSWIFTFormat(swift string) bool {
	if len(swift) != 8 && len(swift) != 11 {
		return false
	}

	swiftRegex := regexp.MustCompile(`^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$`)
	return swiftRegex.MatchString(swift)
}
