package httpgo

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"time"
)

// Response httpgo Response
type Response struct {
	Body []byte

	Request *Request

	HttpRequest  *http.Request
	HttpResponse *http.Response

	Trace *Trace
}

// Trace
type Trace struct {
	DNS          time.Duration
	TLSHandshake time.Duration
	Connect      time.Duration
	Download     time.Duration
	Total        time.Duration
}

// String response string
//
// Usage:
//     resp := httpgo.Response{}
//     resp.String()
func (r Response) String() string {
	return string(r.Body)
}

// Json response json
//
// Usage:
//     type Resp struct {
//        Code   int    `json:"code"`
//        Status string `json:"status"`
//     }
//     res := &Resp{}
//     resp := httpgo.Response{}
//     resp.Json(res)
func (r Response) Json(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// XML response xml
//
// Usage:
//     type Resp struct {
//        Code   int    `xml:"code"`
//        Status string `xml:"status"`
//     }
//     res := &Resp{}
//     resp := httpgo.Response{}
//     resp.XML(res)
func (r Response) XML(v interface{}) error {
	return xml.Unmarshal(r.Body, v)
}
