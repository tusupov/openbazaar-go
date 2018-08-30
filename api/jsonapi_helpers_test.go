package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/OpenBazaar/jsonpb"
	"github.com/OpenBazaar/openbazaar-go/test"
	"github.com/golang/protobuf/proto"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	"os"

	"github.com/op/go-logging"
)

// setupAction is used to change state before and after a set of []apiTest
type setupAction func(*test.Repository) error

// checkFn checks an API response for correctness
type checkFn func(t *testing.T, resp *httptest.ResponseRecorder)

// apiTest is a test case to be run against the api blackbox
type apiTest struct {
	method        string
	path          string
	requestBody   string
	checkResponse checkFn
}

func runAPITests(t *testing.T, tests []apiTest) {
	runAPITestsWithSetup(t, nil, tests)
}

func runAPITestsWithSetup(t *testing.T, setup setupAction, tests []apiTest) {
	gateway, repository, cleanup := setupAPITests(t)
	defer cleanup()

	if setup != nil {
		if err := setup(repository); err != nil {
			t.Fatal("setup:", err)
		}
	}

	for _, test := range tests {
		resp, err := httpRequest(gateway, test.method, test.path, test.requestBody)
		if err != nil {
			t.Fatal(err)
		}

		test.checkResponse(t, resp)
	}
}

// setupAPITests creates a new test environment for api tests
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

func httpRequest(gateway *Gateway, method string, path string, body string) (*httptest.ResponseRecorder, error) {
	// Create a request to the given endpoint
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("test", "test")
	req.AddCookie(test.GetAuthCookie())

	// Send to handler and return response data
	resp := httptest.NewRecorder()
	gateway.handler.ServeHTTP(resp, req)

	return resp, nil
}

func jsonFor(t *testing.T, fixture proto.Message) string {
	m := jsonpb.Marshaler{}

	json, err := m.MarshalToString(fixture)
	if err != nil {
		t.Fatal(err)
	}
	return json
}

func checkIsSuccessJSONEqualTo(expected string) checkFn {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		if resp.Code != 200 {
			t.Fatal("Expected status code 200 but got:", resp.Code)
		}

		var responseJSON interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &responseJSON)
		if err != nil {
			t.Fatal(err)
		}

		var expectedJSON interface{}
		err = json.Unmarshal([]byte(expected), &expectedJSON)
		if err != nil {
			fmt.Println(expected)
			t.Fatal(err)
		}

		if !reflect.DeepEqual(responseJSON, expectedJSON) {
			t.Error("Incorrect response.\nWanted:", expected, "\nGot:", string(resp.Body.Bytes()))
		}
	}
}

func checkIsSuccessJSON(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code != 200 {
		t.Fatal("Expected status code 200 but got:", resp.Code)
	}

	var i interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &i); err != nil {
		t.Fatal("Response is not JSON:", string(resp.Body.Bytes()))
	}
}

func checkIsErrorResponseJSON(errStr string) checkFn {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		var responseJSON struct {
			Success *bool  `json:"success"`
			Reason  string `json:"reason"`
		}

		err := json.Unmarshal(resp.Body.Bytes(), &responseJSON)
		if err != nil {
			t.Fatal(err)
		}

		if responseJSON.Success == nil {
			t.Fatal("success should be false but is not present")
		}

		if *responseJSON.Success {
			t.Fatal("success should be false but is true")
		}

		if !strings.Contains(strings.ToLower(responseJSON.Reason), strings.ToLower(errStr)) {
			t.Fatal(fmt.Sprintf("reason should have '%s' but it does not: %s", errStr, responseJSON.Reason))
		}
	}
}

func checkIsEmptyJSONObject(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code != 200 {
		t.Fatal("Expected status code 200 but got:", resp.Code)
	}

	respBodyStr := string(resp.Body.Bytes())
	if respBodyStr != "{}" {
		t.Fatal("Response is not empty JSON object:", respBodyStr)
	}
}

func checkIsEmptyJSONArray(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code != 200 {
		t.Fatal("Expected status code 200 but got:", resp.Code)
	}

	respBodyStr := string(resp.Body.Bytes())
	if respBodyStr != "[]" {
		t.Fatal("Response is not empty JSON array:", respBodyStr)
	}
}

func checkIs400Error(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code < 400 || resp.Code >= 500 {
		t.Fatal("Expected status code 4XX error but got:", resp.Code)
	}
}

func checkIsNotFoundError(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code != 404 {
		t.Fatal("Expected status code 404 but got:", resp.Code)
	}

	checkIsErrorResponseJSON("not found")(t, resp)
}

func checkIsAlreadyExistsError(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Code != 409 {
		t.Fatal("Expected status code 409 but got:", resp.Code)
	}

	checkIsErrorResponseJSON("already ")(t, resp)
}

func checkIs500Error(err error) checkFn {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		if resp.Code < 500 || resp.Code >= 600 {
			t.Fatal("Expected status code 5XX error but got:", resp.Code)
		}

		checkIsErrorResponseJSON(err.Error())(t, resp)
	}
}
