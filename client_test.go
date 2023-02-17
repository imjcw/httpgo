package httpgo

import (
	"fmt"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	c := &Client{
		BaseUrl:  "https://www.baidu.com",
		Timeout:  5 * time.Second,
		Handlers: []Handler{TraceMiddleware},
	}
	resp, err := c.Get("/?wd=js")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.HttpRequest.Method)
	fmt.Println(string(resp.String()))
	fmt.Println(fmt.Sprintf("%#v", resp.Trace))
}
