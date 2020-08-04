package ghttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

var (
	ErrNoData      = errors.New("no data")
	ErrNotSupport  = errors.New("not support")
	ErrInvalidType = errors.New("invalid type")
)

// NewClient 通过参数创建Client
func NewClient(opts ...Option) *Client {
	o := &Options{}
	o.setNewDefault()
	o.build(opts...)

	client := &http.Client{
		Timeout: o.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   o.DialTimeout,
				KeepAlive: o.KeepAlive,
			}).DialContext,
			TLSHandshakeTimeout: o.HandshakeTimeout,
		},
	}

	c := &Client{client: client, baseURL: o.BaseURL, hooks: o.Hooks}
	return c
}

type Client struct {
	client  *http.Client
	baseURL string
	hooks   []Hook
}

func (c *Client) Get(url string, result interface{}, opts ...Option) (*Response, error) {
	return c.DoRequest(http.MethodGet, url, nil, result, opts...)
}

func (c *Client) Post(url string, req interface{}, result interface{}, opts ...Option) (*Response, error) {
	return c.DoRequest(http.MethodPost, url, req, result, opts...)
}

// DoRequest 执行
func (c *Client) DoRequest(method string, url string, reqBody interface{}, result interface{}, opts ...Option) (*Response, error) {
	o := &Options{}
	o.build(opts...)

	// build url
	if !strings.HasPrefix(url, "http") {
		if o.BaseURL != "" {
			url = path.Join(o.BaseURL, url)
		} else if c.baseURL != "" {
			url = path.Join(o.BaseURL, url)
		}
	}

	body, err := encode(o.ContentType, reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(o.Context, method, url, nil)
	if err != nil {
		return nil, err
	}

	if len(o.Header) > 0 {
		req.Header = o.Header
	}

	if len(o.Query) > 0 {
		req.URL.RawQuery = o.toRawQuery(req.URL.Query())
	}

	addCookies(req, o.Cookies)

	ev := &Event{Req: req, Datas: o.Datas}
	hooks := append(o.Hooks, c.hooks...)

	for i := 0; ; i++ {
		if body != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(body))
		}

		if o.Timeout > 0 {
			ctx, _ := context.WithTimeout(o.Context, o.Timeout)
			req = req.WithContext(ctx)
		}

		ev.SetPrev(i)
		if err := hooks.Run(ev); err != nil {
			return nil, err
		}

		rsp, err := c.client.Do(req)
		ev.SetPost(rsp, err)
		if err := hooks.Run(ev); err != nil {
			return nil, err
		}

		if err == nil {
			if rsp.StatusCode != http.StatusOK {
				return nil, &StatusErr{Code: rsp.StatusCode, Info: rsp.Status}
			}

			if result != nil {
				// decode result
				contentType := o.ContentType
				if val := rsp.Header.Get("Content-Type"); len(val) != 0 {
					contentType = parseContentType(val)
				}
				rspBody, err := ioutil.ReadAll(rsp.Body)
				if err != nil {
					rsp.Body.Close()
					return nil, err
				}
				rsp.Body.Close()
				rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

				if err := decode(contentType, rspBody, result); err != nil {
					return nil, err
				}
			}

			return rsp, nil
		} else if isTimeoutErr(err) && i < o.Retry {
			wait := o.Backoff.Next()
			for {
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(wait):
					break
				}
			}
		} else {
			return nil, err
		}
	}
}

// isTimeoutErr 判断是否是超时错误
func isTimeoutErr(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}

	return false
}

// StatusErr 当Response返回状态非200时,返回此错误
type StatusErr struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

func (se *StatusErr) Error() string {
	return fmt.Sprintf("invalid http status,code=%+v, info=%+v", se.Code, se.Info)
}

// IsStatusErr 判断是否是StatusErr错误
func IsStatusErr(e error) bool {
	_, ok := e.(*StatusErr)
	return ok
}
