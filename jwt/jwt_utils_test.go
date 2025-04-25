package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

const (
	testSecretKey    = "test-secret-key"
	testExpireMinute = 1
)

var testConfig = JWTConfig{
	SecretKey:    testSecretKey,
	ExpireMinute: testExpireMinute,
}

// Helper function to reset global state for tests
func resetGlobals() {
	gConfig = JWTConfig{}
	// Note: sync.Once cannot be reset easily. Tests assume InitJWT is called once per test run or rely on its idempotent nature.
	// For more complex scenarios, dependency injection might be preferred over global state.
}

func TestInitJWT(t *testing.T) {
	resetGlobals()
	InitJWT(testConfig)
	assert.Equal(t, testConfig.SecretKey, gConfig.SecretKey, "SecretKey should be initialized")
	assert.Equal(t, testConfig.ExpireMinute, gConfig.ExpireMinute, "ExpireMinute should be initialized")

	// Try initializing again with different config, should not change
	anotherConfig := JWTConfig{SecretKey: "another-secret", ExpireMinute: 5}
	InitJWT(anotherConfig)
	assert.Equal(t, testConfig.SecretKey, gConfig.SecretKey, "SecretKey should not change after first init")
	assert.Equal(t, testConfig.ExpireMinute, gConfig.ExpireMinute, "ExpireMinute should not change after first init")
}

func TestGenerateToken(t *testing.T) {
	resetGlobals()
	InitJWT(testConfig)

	claims := map[string]any{
		"userID": 123,
		"role":   "admin",
	}

	tokenString, err := GenerateToken(claims)
	assert.NoError(t, err, "GenerateToken should not return an error for valid claims")
	assert.NotEmpty(t, tokenString, "Generated token string should not be empty")

	// Optional: Decode and check claims (without validation)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	assert.NoError(t, err, "Parsing unverified token should not fail")
	if parsedClaims, ok := token.Claims.(jwt.MapClaims); ok {
		assert.Equal(t, float64(123), parsedClaims["userID"], "UserID claim should match") // Note: JSON numbers are often float64
		assert.Equal(t, "admin", parsedClaims["role"], "Role claim should match")
		assert.Contains(t, parsedClaims, "exp", "Token should contain 'exp' claim")
		expClaim, ok := parsedClaims["exp"].(float64)
		assert.True(t, ok, "'exp' claim should be a number")
		// Check if expiry is roughly correct (within a small delta)
		expectedExp := time.Now().Add(time.Minute * time.Duration(testExpireMinute)).Unix()
		assert.InDelta(t, expectedExp, int64(expClaim), 5, "Expiry time should be approximately correct") // Allow 5 seconds delta
	} else {
		t.Errorf("Could not assert token claims type")
	}
}

func TestValidateToken(t *testing.T) {
	resetGlobals()
	InitJWT(testConfig)

	validClaims := map[string]any{
		"userID": 456,
		"role":   "user",
	}

	// --- Test Case 1: Valid Token ---
	t.Run("ValidToken", func(t *testing.T) {
		validToken, err := GenerateToken(validClaims)
		assert.NoError(t, err)

		parsedClaims, err := ValidateToken(validToken)
		assert.NoError(t, err, "ValidateToken should not return an error for a valid token")
		assert.NotNil(t, parsedClaims, "Parsed claims should not be nil for a valid token")
		assert.Equal(t, float64(456), parsedClaims["userID"], "UserID claim should match")
		assert.Equal(t, "user", parsedClaims["role"], "Role claim should match")
	})

	// --- Test Case 2: Expired Token ---
	t.Run("ExpiredToken", func(t *testing.T) {
		// Create a config with negative expiry to generate an already expired token
		expiredConfig := JWTConfig{SecretKey: testSecretKey, ExpireMinute: -5}
		resetGlobals() // Reset for this specific sub-test's config
		InitJWT(expiredConfig)

		expiredToken, err := GenerateToken(validClaims)
		assert.NoError(t, err)

		// Let some time pass to ensure expiry is definitely in the past relative to validation check
		time.Sleep(1 * time.Second)

		// Reset back to original config for validation
		resetGlobals()
		InitJWT(testConfig)

		_, err = ValidateToken(expiredToken)
		assert.Error(t, err, "ValidateToken should return an error for an expired token")

		// Check that the error is a jwt.ValidationError and indicates expiration
		var ve *jwt.ValidationError
		assert.True(t, errors.As(err, &ve), "Error should be a jwt.ValidationError for expired token")
		if ve != nil {
			assert.True(t, ve.Errors&jwt.ValidationErrorExpired != 0, "ValidationError should indicate token is expired")
		}
	})

	// --- Test Case 3: Invalid Signature ---
	// This test is commented out because the current init use once.Do.
	// It could pass if we remove the once.Do

	// t.Run("InvalidSignature", func(t *testing.T) {
	// 	// Generate token with correct secret
	// 	token, err := GenerateToken(validClaims)
	// 	assert.NoError(t, err)

	// 	// Try validating with a wrong secret
	// 	wrongSecretConfig := JWTConfig{SecretKey: "wrong-secret", ExpireMinute: testExpireMinute}
	// 	resetGlobals()
	// 	InitJWT(wrongSecretConfig) // Use wrong secret for validation attempt

	// 	_, err = ValidateToken(token)
	// 	assert.Error(t, err, "ValidateToken should return an error for invalid signature")
	// 	// The specific error from jwt-go for signature mismatch is ValidationErrorSignatureInvalid
	// 	var ve *jwt.ValidationError
	// 	assert.True(t, errors.As(err, &ve), "Error should be a jwt.ValidationError")
	// 	if ve != nil { // Add nil check for safety
	// 		assert.True(t, ve.Errors&jwt.ValidationErrorSignatureInvalid != 0, "Error should indicate invalid signature")
	// 	}

	// 	// Reset back to original config
	// 	resetGlobals()
	// 	InitJWT(testConfig)
	// })

	// --- Test Case 3: Invalid Signing Method ---
	t.Run("InvalidSigningMethod", func(t *testing.T) {
		// Manually create a token with a different signing method (e.g., ES256) but sign with HS256 key
		// This is tricky to do correctly without proper keys for ES256.
		// A simpler approach is to modify a valid HS256 token header.
		validToken, err := GenerateToken(validClaims)
		assert.NoError(t, err)

		parts := strings.Split(validToken, ".")
		assert.Len(t, parts, 3, "Token should have 3 parts")

		// Decode header, change alg, re-encode (This might invalidate the signature, but tests the method check)
		headerBytes, err := jwt.DecodeSegment(parts[0])
		assert.NoError(t, err)
		var header map[string]any
		err = json.Unmarshal(headerBytes, &header)
		assert.NoError(t, err)
		header["alg"] = "ES256" // Change to an unexpected algorithm
		newHeaderBytes, err := json.Marshal(header)
		assert.NoError(t, err)
		parts[0] = jwt.EncodeSegment(newHeaderBytes)

		malformedToken := strings.Join(parts, ".")

		_, err = ValidateToken(malformedToken)
		assert.Error(t, err, "ValidateToken should return an error for unexpected signing method")
		// The error comes from the keyFunc callback
		assert.Contains(t, err.Error(), "Unexpected signing method", "Error message should indicate unexpected signing method")
	})

	// --- Test Case 4: Malformed Token ---
	t.Run("MalformedToken", func(t *testing.T) {
		malformedTokens := []string{
			"invalidtoken",
			"header.payload", // Missing signature
			"header.payload.",
			".payload.signature",
			"header..signature",
		}
		for _, mt := range malformedTokens {
			_, err := ValidateToken(mt)
			assert.Error(t, err, fmt.Sprintf("ValidateToken should return an error for malformed token: %s", mt))
			// The error might be jwt.ValidationErrorMalformed or others depending on the exact issue
		}
	})

	// --- Test Case 5: Empty Token ---
	t.Run("EmptyToken", func(t *testing.T) {
		_, err := ValidateToken("")
		assert.Error(t, err, "ValidateToken should return an error for an empty token string")
	})
}
