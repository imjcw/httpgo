package httpgo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

// Client .
//
// Handlers是按正序执行的
// NoProxy为true时, Proxy失效; NoProxy为false, 且Proxy未配置, 将使用系统默认的Proxy
// IgnoreSSL为true时, 将忽略https的验证
type Client struct {
	BaseUrl string

	UserAgent string
	Timeout   time.Duration
	Proxy     string
	NoProxy   bool
	IgnoreSSL bool

	Handlers []Handler

	client http.Client
}

var DefaultClient = &Client{
	client: http.Client{},
}

// Get  net/http Get
func (c *Client) Get(uri string) (resp *Response, err error) {
	return c.Do(&Request{Uri: uri})
}

// Post net/http Post
func (c *Client) Post(uri string, contentType string, body io.Reader) (resp *Response, err error) {
	return c.Do(&Request{
		Uri: uri,
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: body,
	})
}

// PostJson post json
func (c *Client) PostJson(uri string, body interface{}) (*Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.Do(&Request{
		Uri: uri,
		Header: http.Header{
			"Content-Type": []string{ContentJson},
		},
		Body: bytes.NewReader(payload),
	})
}

// PostForm post form
func (c *Client) PostForm(uri string, body url.Values) (*Response, error) {
	payload := []byte(body.Encode())
	return c.Do(&Request{
		Uri: uri,
		Header: http.Header{
			"Content-Type": []string{ContentForm},
		},
		Body: bytes.NewReader(payload),
	})
}

// Do send request
func (c Client) Do(req *Request) (resp *Response, err error) {
	defer errRecover()
	ctx := &Context{
		Request:  req,
		Response: &Response{},
	}
	transHandlers(ctx, c)
	transRequest(req)
	reqUrl := strings.TrimRight(c.BaseUrl, "/") + "/" + strings.TrimLeft(req.Uri, "/")
	ctx.HttpRequest, err = http.NewRequestWithContext(req.Context, req.Method, reqUrl, req.Body)
	if err != nil {
		return nil, err
	}
	transHeader(req, ctx.HttpRequest, c)
	transQuery(req, ctx.HttpRequest)
	client := transRuntimeOption(c, req.RuntimeOption)
	ctx.handlers = append(ctx.handlers, func(c *Context) {
		ctx.Response.HttpResponse, err = client.Do(ctx.HttpRequest)
		ctx.Response.Request = ctx.Request
		ctx.Response.HttpRequest = ctx.HttpRequest
		ctx.HttpResponse = ctx.Response.HttpResponse
		if err == nil {
			ctx.Response.Body, err = ioutil.ReadAll(ctx.Response.HttpResponse.Body)
			if err == nil {
				ctx.Response.HttpResponse.Body.Close()
			}
		}
		ctx.Next()
	})
	ctx.handlers[0](ctx)
	return ctx.Response, nil
}

func errRecover() error {
	e := recover()
	if e == nil {
		return nil
	}
	if er, ok := e.(error); ok {
		return er
	}
	return errors.New(fmt.Sprintf("%v", e))
}

func transHandlers(ctx *Context, c Client) {
	for _, handler := range c.Handlers {
		ctx.handlers = append(ctx.handlers, handler)
	}
}

func transRequest(req *Request) {
	if req.Context == nil {
		req.Context = context.Background()
	}
}

func transHeader(req *Request, newReq *http.Request, c Client) {
	for k, values := range req.Header {
		for _, v := range values {
			newReq.Header.Set(k, v)
		}
	}
	userAgent := "HttpGO " + runtime.GOOS + " " + runtime.Version()
	if len(c.UserAgent) > 0 {
		userAgent = c.UserAgent
	}
	newReq.Header.Set("User-Agent", userAgent)
}

func transQuery(req *Request, newReq *http.Request) {
	newUrl := url.URL{}
	newUrl.Path = newReq.URL.Path
	query := newReq.URL.Query()
	for k, v := range req.Query {
		query.Set(k, v)
	}
	newReq.URL, _ = newReq.URL.Parse(newUrl.String() + "?" + query.Encode())
}

func transRuntimeOption(c Client, option RuntimeOption) *http.Client {
	ignoreSSL := c.IgnoreSSL
	if option.IgnoreSSL {
		ignoreSSL = true
	}
	tr := &http.Transport{
		Proxy: transProxy(c, option),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: ignoreSSL,
		},
	}
	c.client.Transport = tr
	if c.Timeout > 0 {
		c.client.Timeout = c.Timeout
	}
	if option.Timeout > 0 {
		c.client.Timeout = option.Timeout
	}
	return &c.client
}

func transProxy(c Client, option RuntimeOption) (proxy func(*http.Request) (*url.URL, error)) {
	if option.NoProxy {
		return
	}
	if len(option.Proxy) > 0 {
		return func(_ *http.Request) (*url.URL, error) {
			return url.Parse(option.Proxy)
		}
	}
	if c.NoProxy {
		return
	}
	if len(c.Proxy) > 0 {
		return func(_ *http.Request) (*url.URL, error) {
			return url.Parse(c.Proxy)
		}
	}
	return http.ProxyFromEnvironment
}
