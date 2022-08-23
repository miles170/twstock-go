package twstock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-querystring/query"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

const (
	Version = "1.0.0"

	defaultTwseBaseURL     = "https://www.twse.com.tw"
	defaultTpexBaseURL     = "https://www.tpex.org.tw"
	defaultMisTwseBaseURL  = "https://mis.twse.com.tw/"
	defaultIsinTwseBaseURL = "https://isin.twse.com.tw"
)

// A Client manages communication with the API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	twseBaseURL *url.URL
	twseDecoder transform.Transformer

	tpexBaseURL *url.URL
	tpexDecoder transform.Transformer

	misTwseBaseURL *url.URL

	isinTwseBaseURL *url.URL
	isinTwseDecoder transform.Transformer

	// Services used for talking to different parts of the API.
	MarketData *MarketDataService
	Security   *SecurityService
	Quote      *QuoteService
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(u *url.URL, opts interface{}) (*url.URL, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return u, nil
	}

	qs, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	u.RawQuery = qs.Encode()
	return u, nil
}

// NewClient returns a new Fugle API client.
func NewClient() *Client {
	httpClient := &http.Client{}
	twseBaseURL, _ := url.Parse(defaultTwseBaseURL)
	tpexBaseURL, _ := url.Parse(defaultTpexBaseURL)
	misTwseBaseURL, _ := url.Parse(defaultMisTwseBaseURL)
	isinTwseBaseURL, _ := url.Parse(defaultIsinTwseBaseURL)
	c := &Client{
		client: httpClient,

		twseBaseURL: twseBaseURL,
		twseDecoder: transform.Nop,

		tpexBaseURL: tpexBaseURL,
		tpexDecoder: transform.Nop,

		misTwseBaseURL: misTwseBaseURL,

		isinTwseBaseURL: isinTwseBaseURL,
		isinTwseDecoder: traditionalchinese.Big5.NewDecoder(),
	}
	c.MarketData = &MarketDataService{client: c}
	c.Security = &SecurityService{client: c}
	c.Quote = &QuoteService{client: c}
	return c
}

// NewRequest creates an API request.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	contentType := ""

	if v, ok := body.(string); ok {
		buf = strings.NewReader(v)
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// Do sends an API request and returns the API response.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

// Do sends an API request and returns the goquery.Document.
func (c *Client) DoTransformToDocument(req *http.Request, t transform.Transformer) (*goquery.Document, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	body := transform.NewReader(resp.Body, t)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode)
}

// CheckResponse checks the API response for errors
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	return errorResponse
}
