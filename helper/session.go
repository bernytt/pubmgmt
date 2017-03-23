package helper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func checkHTTPRespStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case 200, 201, 202, 204, 206:
		return nil
	case 400:
		return errors.New("Error: 400 BadRequest")
	case 401:
		return errors.New("Error: 403 Unauthorised")
	case 403:
		return errors.New("Error: 403 Forbidden")
	case 404:
		return errors.New("Error: 404 NotFound")
	case 405:
		return errors.New("Error: 405 MethodNotAllowed")
	case 409:
		return errors.New("Error: 409 Conflict")
	case 413:
		return errors.New("Error: 413 OverLimit")
	case 415:
		return errors.New("Error: 415 BadMediaType")
	case 422:
		return errors.New("Error: 422 Unprocessable")
	case 429:
		return errors.New("Error: 429 TooManyRequest")
	case 500:
		return errors.New("Error: 500 InternalServerError")
	case 501:
		return errors.New("Error: 501 NotImplemented")
	case 503:
		return errors.New("Error: 503 ServiceUnavailable")
	default:
		return errors.New("Error: Unexpected response status code")
	}
}

type Response struct {
	Resp *http.Response
	Body []byte
}

type Session struct {
	cli     *http.Client
	Headers http.Header
}

func NewSession(cli *http.Client, tls *tls.Config) (session *Session) {
	if cli == nil {
		tr := &http.Transport{
			TLSClientConfig:    tls,
			DisableCompression: true,
		}
		cli = &http.Client{Transport: tr}
	}
	session = &Session{
		cli:     cli,
		Headers: http.Header{},
	}
	return session
}

func (s *Session) NewRequest(method, url string, headers *http.Header, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		req.Header = *headers
	}
	return
}

func (s *Session) Do(req *http.Request) (resp *http.Response, err error) {
	for k := range s.Headers {
		req.Header.Set(k, s.Headers.Get(k))
	}
	resp, err = s.cli.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Session) Request(method, url string, params *url.Values, headers *http.Header, body *[]byte) (resp *http.Response, err error) {
	if params != nil {
		url = url + "?" + params.Encode()
	}
	var buf io.Reader
	if body != nil {
		buf = bytes.NewReader(*body)
	}
	req, err := s.NewRequest(method, url, headers, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err = s.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s *Session) RequestJSON(method, url string, params *url.Values, headers *http.Header, body interface{}, responseContainer interface{}) (resp *http.Response, err error) {
	bodyjson, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = &http.Header{}
		headers.Add("Accept", "application/json")
	}
	resp, err = s.Request(method, url, params, headers, &bodyjson)
	if err != nil {
		return nil, err
	}
	err = checkHTTPRespStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error reading response body")
	}
	if err = json.Unmarshal(rbody, &responseContainer); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Session) Delete(url string, params *url.Values, headers *http.Header) (resp *http.Response, err error) {
	return s.Request("DELETE", url, params, headers, nil)
}

func (s *Session) Get(url string, params *url.Values, headers *http.Header) (resp *http.Response, err error) {
	return s.Request("GET", url, params, headers, nil)
}

func (s *Session) GetJSON(url string, params *url.Values, headers *http.Header, responseContainer interface{}) (resp *http.Response, err error) {
	return s.RequestJSON("GET", url, params, headers, nil, responseContainer)
}

func (s *Session) Head(url string, params *url.Values, headers *http.Header) (resp *http.Response, err error) {
	return s.Request("HEAD", url, params, headers, nil)
}

func (s *Session) Post(url string, params *url.Values, headers *http.Header, body *[]byte) (resp *http.Response, err error) {
	return s.Request("POST", url, params, headers, body)
}

func (s *Session) PostJSON(url string, params *url.Values, headers *http.Header, body, responseContainer interface{}) (resp *http.Response, err error) {
	return s.RequestJSON("POST", url, params, headers, body, responseContainer)
}

func (s *Session) Put(url string, params *url.Values, headers *http.Header, body *[]byte) (resp *http.Response, err error) {
	return s.Request("PUT", url, params, headers, body)
}
