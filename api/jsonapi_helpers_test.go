package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/OpenBazaar/jsonpb"
	"github.com/OpenBazaar/openbazaar-go/test"
	"github.com/golang/protobuf/proto"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	"os"

	"github.com/op/go-logging"
)

// testURIRoot is the root http URI to hit for testing
const testURIRoot = "http://127.0.0.1:9191"

// anyResponseJSON is a sentinel denoting any valid JSON response body is valid
const anyResponseJSON = "__anyresponsebodyJSON__"

// testHTTPClient is the http client to use for tests
var testHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

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

// apiTests is a slice of apiTest
type apiTests []apiTest

func runAPITests(t *testing.T, tests apiTests) {
	// _, err := test.ResetRepository()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	gateway, _, cleanup := setupAPITests(t)
	defer cleanup()

	for _, jsonAPITest := range tests {
		executeAPITest(t, gateway.handler, jsonAPITest)
	}
}

func runAPITestsWithSetup(t *testing.T, tests apiTests, runBefore, runAfter setupAction) {
	// repository, err := test.ResetRepository()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	gateway, repository, cleanup := setupAPITests(t)
	defer cleanup()

	if runBefore != nil {
		if err := runBefore(gateway, repository); err != nil {
			t.Fatal("runBefore:", err)
		}
	}

	for _, jsonAPITest := range tests {
		executeAPITest(t, gateway.handler, jsonAPITest)
	}

	if runAfter != nil {
		if err := runAfter(gateway, repository); err != nil {
			t.Fatal("runAfter:", err)
		}
	}
}

func runAPITest(t *testing.T, subject apiTest) {
	// _, err := test.ResetRepository()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	gateway, _, cleanup := setupAPITests(t)
	defer cleanup()
	executeAPITest(t, gateway.handler, subject)
}

// executeAPITest executes the given test against the blackbox
func executeAPITest(t *testing.T, mux http.Handler, test apiTest) {
	// Make the request
	req, err := buildRequest(test.method, test.path, test.requestBody)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	// mux, ok := l.(http.ServeMux)
	// if !ok {
	// 	t.Fatal("blah")
	// }
	mux.ServeHTTP(resp, req)

	// resp, err := testHTTPClient.Do(req)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// Ensure correct status code
	if resp.Code != test.expectedResponseCode {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(test.method, test.path, string(b))
		t.Errorf("Wanted status %d, got %d", test.expectedResponseCode, resp.Code)
		return
	}

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Parse response as JSON
	var responseJSON interface{}
	err = json.Unmarshal(respBody, &responseJSON)
	if err != nil {
		t.Fatal(err)
	}

	// Unless explicitly saying any JSON is expected check for equality
	if test.expectedResponseBody != anyResponseJSON {
		var expectedJSON interface{}
		err = json.Unmarshal([]byte(test.expectedResponseBody), &expectedJSON)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(responseJSON, expectedJSON) {
			fmt.Println("expected:", test.expectedResponseBody)
			fmt.Println("actual:", string(respBody))
			t.Error("Incorrect response")
		}
	}
}

// buildRequest issues an http request directly to the blackbox handler
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

func errorResponseJSON(err error) string {
	return `{"success": false, "reason": "` + err.Error() + `"}`
}

func httpGet(gateway *Gateway, endpoint string) ([]byte, error) {
	req, err := buildRequest("GET", endpoint, "")
	if err != nil {
		return nil, err
	}
	resp := httptest.NewRecorder()

	gateway.handler.ServeHTTP(resp, req)

	if resp.Code != 200 {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func jsonFor(t *testing.T, fixture proto.Message) string {
	m := jsonpb.Marshaler{}

	json, err := m.MarshalToString(fixture)
	if err != nil {
		t.Fatal(err)
	}
	return json
}
