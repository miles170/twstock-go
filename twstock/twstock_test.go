package twstock

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/text/transform"
)

// setup sets up a test HTTP server along with a twstock.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *Client, mux *http.ServeMux, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the fugle client being tested and is
	// configured to use test server.
	client = NewClient()
	url, _ := url.Parse(server.URL + "/")
	client.isinTwseBaseURL = url

	return client, mux, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testErrorContains(t *testing.T, e error, want string) {
	t.Helper()
	if !strings.Contains(e.Error(), want) {
		t.Errorf("testErrorContains: err message = %s, want %s", e.Error(), want)
	}
}

func testURLParseError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func TestNewRequest_BadURL(t *testing.T) {
	c := NewClient()
	_, err := c.NewRequest("GET", ":")
	testURLParseError(t, err)
}

func TestNewRequest_BadMethod(t *testing.T) {
	c := NewClient()
	if _, err := c.NewRequest("BOGUS\nMETHOD", "."); err == nil {
		t.Fatal("NewRequest returned nil; expected error")
	}
}

func TestDoTransformToDocument_BadRequestURL(t *testing.T) {
	client, _, teardown := setup()
	defer teardown()

	req, err := client.NewRequest("GET", "test-url")
	if err != nil {
		t.Fatalf("client.NewRequest returned error: %v", err)
	}

	resp, err := client.DoTransformToDocument(req, nil)
	if resp != nil {
		t.Errorf("client.DoTransformToDocument resp = %#v, want nil", resp)
	}
	if err == nil {
		t.Error("client.DoTransformToDocument err = nil, want error")
	}
}

type errDecoder struct{ transform.NopResetter }

func (errDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	return nDst, nSrc, transform.ErrShortDst
}

func TestDoTransformToDocument_BadDocument(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/test-url", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	u, err := client.isinTwseBaseURL.Parse("/test-url")
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}

	req, err := client.NewRequest("GET", u.String())
	if err != nil {
		t.Fatalf("client.NewRequest returned error: %v", err)
	}

	resp, err := client.DoTransformToDocument(req, errDecoder{})
	if resp != nil {
		t.Errorf("client.DoTransformToDocument resp = %#v, want nil", resp)
	}
	if err == nil {
		t.Error("client.DoTransformToDocument err = nil, want error")
	}
}
