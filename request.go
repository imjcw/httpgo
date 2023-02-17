package httpgo

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Request httpgo request
type Request struct {
	Method string
	Header http.Header
	Uri    string
	Query  map[string]string
	Body   io.Reader

	Context       context.Context
	RuntimeOption RuntimeOption
}

// RuntimeOption .
type RuntimeOption struct {
	Timeout   time.Duration
	Proxy     string
	NoProxy   bool
	IgnoreSSL bool
}
