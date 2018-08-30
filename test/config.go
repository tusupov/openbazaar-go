package test

import (
	"math"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/OpenBazaar/openbazaar-go/schema"
)

// NewAPIConfig returns a new config object for the API tests
func NewAPIConfig() (*schema.APIConfig, error) {

	apiConfig := &schema.APIConfig{
		Enabled:       true,
		Authenticated: true,
		Username:      "test",
		Password:      "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", // sha256("test")
	}

	return apiConfig, nil
}

// GetAuthCookie returns a pointer to a test authentication cookie
func GetAuthCookie() *http.Cookie {
	return &http.Cookie{
		Name:  "OpenBazaar_Auth_Cookie",
		Value: "supersecret",
	}
}

// getNewRepoPath a new repo path to use for tests
func getNewRepoPath() string {
	base := getEnvString("OPENBAZAAR_TEST_REPO_PATH", "/tmp/openbazaar-test")
	return path.Join(base, strconv.FormatInt(rand.Int63n(math.MaxInt64), 16))
}

// getMnemonic returns a static mnemonic to use
func getMnemonic() string {
	return getEnvString("OPENBAZAAR_TEST_MNEMONIC", "correct horse battery staple")
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
