package veracode

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	hmacAlgorithm  = "VERACODE-HMAC-SHA-256"
	requestVersion = "vcode_request_version_1"
)

// generateHMACHeader builds the Veracode Authorization header.
// See: https://docs.veracode.com/r/c_api_signing
func generateHMACHeader(apiID, apiSecret string, request *http.Request) (string, error) {
	secretKey, err := hex.DecodeString(removeCredentialPrefix(apiSecret))
	if err != nil {
		return "", fmt.Errorf("invalid API secret: %w", err)
	}

	apiID = removeCredentialPrefix(apiID)

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	requestURL := request.URL.EscapedPath()
	if request.URL.RawQuery != "" {
		requestURL += "?" + request.URL.RawQuery
	}

	data := fmt.Sprintf(
		"id=%s&host=%s&url=%s&method=%s",
		apiID,
		request.URL.Host,
		requestURL,
		request.Method,
	)

	signature := generateSignature(secretKey, nonceBytes, timestamp, data)

	return fmt.Sprintf(
		"%s id=%s,ts=%s,nonce=%s,sig=%s",
		hmacAlgorithm,
		apiID,
		timestamp,
		nonce,
		signature,
	), nil
}

func generateSignature(secretKey, nonce []byte, timestamp, data string) string {
	encryptedNonce := hmacSHA256(nonce, secretKey)
	encryptedTimestamp := hmacSHA256([]byte(timestamp), encryptedNonce)
	signingKey := hmacSHA256([]byte(requestVersion), encryptedTimestamp)
	signature := hmacSHA256([]byte(data), signingKey)

	return hex.EncodeToString(signature)
}

func hmacSHA256(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

// removeCredentialPrefix strips the "<prefix>-" part of API id/secret
// values (e.g. "abcdef12-..."), if present.
func removeCredentialPrefix(value string) string {
	parts := strings.SplitN(value, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return value
}
