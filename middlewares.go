package httpgo

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

type Handler func(c *Context)

// Context .
type Context struct {
	Request      *Request
	Response     *Response
	HttpRequest  *http.Request
	HttpResponse *http.Response

	handlers []Handler
	index    int8
}

// Next .
func (c *Context) Next() {
	c.index++
	if int8(len(c.handlers)) > c.index {
		c.handlers[c.index](c)
		c.index++
	}
}

// TraceMiddleware .
func TraceMiddleware(c *Context) {
	var dnsStart, tlsHandshakeStart, connectStart, downloadStart time.Time
	var startTime = time.Now()
	if c.Response.Trace == nil {
		c.Response.Trace = &Trace{}
	}
	c.HttpRequest = c.HttpRequest.WithContext(httptrace.WithClientTrace(c.HttpRequest.Context(), &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(ddi httptrace.DNSDoneInfo) { c.Response.Trace.DNS = time.Since(dnsStart) },

		TLSHandshakeStart: func() { tlsHandshakeStart = time.Now() },
		TLSHandshakeDone:  func(cs tls.ConnectionState, err error) { c.Response.Trace.TLSHandshake = time.Since(tlsHandshakeStart) },

		ConnectStart: func(network, addr string) { connectStart = time.Now() },
		ConnectDone:  func(network, addr string, err error) { c.Response.Trace.Connect = time.Since(connectStart) },

		GotFirstResponseByte: func() { downloadStart = time.Now() },
	}))
	c.Next()
	c.Response.Trace.Download = time.Since(downloadStart)
	c.Response.Trace.Total = time.Since(startTime)
}
