package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/OpenBazaar/jsonpb"
	"github.com/OpenBazaar/openbazaar-go/test"
	"github.com/golang/protobuf/proto"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	"os"

	"github.com/op/go-logging"
)

// anyResponseJSON is a sentinel denoting any valid JSON response body is valid
const anyResponseJSON = "__anyresponsebodyJSON__"

// setupAPITests starts a new API gateway listening on the default test interface
func setupAPITests(t *testing.T) (*Gateway, *test.Repository, func()) {
	// Create test repo
	repository, err := test.NewRepository()
	if err != nil {
		t.Fatal(err)
	}

	// Create a test node, cookie, and config
	node, err := test.NewNode(repository)
	if err != nil {
		t.Fatal(err)
	}

	apiConfig, err := test.NewAPIConfig()
	if err != nil {
		t.Fatal(err)
	}

	addr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatal(err)
	}

	listener, err := manet.Listen(addr)
	if err != nil {
		t.Fatal(err)
	}

	gateway, err := NewGateway(node, *test.GetAuthCookie(), listener.NetListener(), *apiConfig, logging.NewLogBackend(os.Stdout, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	return gateway, repository, func() {
		err := repository.Delete()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// apiTest is a test case to be run against the api blackbox
type apiTest struct {
	method      string
	path        string
	requestBody string

	expectedResponseCode int
	expectedResponseBody string
}

// setupAction is used to change state before and after a set of []apiTest
type setupAction func(*Gateway, *test.Repository) error

func runAPITests(t *testing.T, tests []apiTest) {
	gateway, _, cleanup := setupAPITests(t)
	defer cleanup()

	for _, jsonAPITest := range tests {
		executeAPITest(t, gateway, jsonAPITest)
	}
}

func runAPITestsWithSetup(t *testing.T, tests []apiTest, runBefore, runAfter setupAction) {
	gateway, repository, cleanup := setupAPITests(t)
	defer cleanup()

	if runBefore != nil {
		if err := runBefore(gateway, repository); err != nil {
			t.Fatal("runBefore:", err)
		}
	}

	for _, jsonAPITest := range tests {
		executeAPITest(t, gateway, jsonAPITest)
	}

	if runAfter != nil {
		if err := runAfter(gateway, repository); err != nil {
			t.Fatal("runAfter:", err)
		}
	}
}

// executeAPITest executes the given test against the blackbox
func executeAPITest(t *testing.T, gateway *Gateway, test apiTest) {
	// Make request
	respBody, respCode, err := httpRequest(gateway, test.method, test.path, test.requestBody)
	if err != nil {
		t.Fatal(err)
	}

	// Parse response as JSON
	var responseJSON interface{}
	err = json.Unmarshal(respBody, &responseJSON)
	if err != nil {
		t.Fatal(err)
	}

	// Assert correctness
	if test.expectedResponseCode != respCode {
		t.Fatal("Incorrect response code.\nWanted:", test.expectedResponseCode, "\nGot:", respCode)
	}

	if test.expectedResponseBody != anyResponseJSON {
		var expectedJSON interface{}
		err = json.Unmarshal([]byte(test.expectedResponseBody), &expectedJSON)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(responseJSON, expectedJSON) {
			t.Error("Incorrect response.\nWanted:", test.expectedResponseBody, "\nGot:", string(respBody))
		}
	}
}

// buildRequest creates an api request
func buildRequest(method string, path string, body string) (*http.Request, error) {
	// Create a JSON request to the given endpoint
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	// Set headers/auth/cookie
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("test", "test")
	req.AddCookie(test.GetAuthCookie())

	return req, nil
}

func httpRequest(gateway *Gateway, method string, endpoint string, body string) ([]byte, int, error) {
	req, err := buildRequest(method, endpoint, body)
	if err != nil {
		return nil, 0, err
	}
	resp := httptest.NewRecorder()

	gateway.handler.ServeHTTP(resp, req)

	return resp.Body.Bytes(), resp.Code, nil
}

func jsonFor(t *testing.T, fixture proto.Message) string {
	m := jsonpb.Marshaler{}

	json, err := m.MarshalToString(fixture)
	if err != nil {
		t.Fatal(err)
	}
	return json
}

func errorResponseJSON(err error) string {
	return `{"success": false, "reason": "` + err.Error() + `"}`
}
