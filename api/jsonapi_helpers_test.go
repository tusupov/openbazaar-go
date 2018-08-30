package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

type checkFn func(t *testing.T, respBody []byte)

// apiTest is a test case to be run against the api blackbox
type apiTest struct {
	method      string
	path        string
	requestBody string

	expectedResponseCode int
	checkResponseFn      checkFn
}

func runAPITests(t *testing.T, tests []apiTest) {
	runAPITestsWithSetup(t, nil, tests)
}

func runAPITestsWithSetup(t *testing.T, runBefore setupAction, tests []apiTest) {
	gateway, repository, cleanup := setupAPITests(t)
	defer cleanup()

	if runBefore != nil {
		if err := runBefore(repository); err != nil {
			t.Fatal("runBefore:", err)
		}
	}

	for _, test := range tests {
		// Make request
		respBody, respCode, err := httpRequest(gateway, test.method, test.path, test.requestBody)
		if err != nil {
			t.Fatal(err)
		}

		// Assert correctness
		if test.expectedResponseCode != respCode {
			t.Fatal("Incorrect response code.\nWanted:", test.expectedResponseCode, "\nGot:", respCode)
		}

		test.checkResponseFn(t, respBody)
	}
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

func httpRequest(gateway *Gateway, method string, path string, body string) ([]byte, int, error) {
	// Create a JSON request to the given endpoint
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		return nil, 0, err
	}

	// Set headers/auth/cookie
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("test", "test")
	req.AddCookie(test.GetAuthCookie())

	// Send to handler and return response data
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
