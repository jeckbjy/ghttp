package ghttp

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const (
	TypeJSON = "application/json"
	TypeXML  = "application/xml"
	TypeForm = "application/x-www-form-urlencoded"
	TypeHTML = "text/html"
	TypeText = "text/plain"
)

const (
	UTF8 = "utf-8"
)

const (
	defaultTimeout          = time.Second * 60
	defaultDialTimeout      = time.Second * 60
	defaultKeepAlive        = time.Second * 60
	defaultHandshakeTimeout = time.Second * 60
	defaultContentType      = TypeJSON
)

var defaultBackoff = NewConstantBackoff(time.Second)

type Request = http.Request
type Response = http.Response

type EventType int

const (
	EventPrev = EventType(0)
	EventPost = EventType(1)
)

type Event struct {
	Type  EventType
	Req   *Request
	Rsp   *Response         //
	Err   error             //
	Num   int               // 执行次数
	Datas map[string]string // 扩展参数，由Options传过来
}

func (ev *Event) SetPrev(num int) {
	ev.Type = EventPrev
	ev.Num = num
}

func (ev *Event) SetPost(rsp *Response, err error) {
	ev.Type = EventPost
	ev.Rsp = rsp
	ev.Err = err
}

type Hook func(ev *Event) error
type Hooks []Hook

func (hooks Hooks) Run(ev *Event) error {
	for _, h := range hooks {
		if err := h(ev); err != nil {
			return err
		}
	}

	return nil
}

type Option func(o *Options)
type Options struct {
	Context          context.Context   //
	BaseURL          string            //
	Timeout          time.Duration     // 超时时间
	DialTimeout      time.Duration     //
	HandshakeTimeout time.Duration     //
	KeepAlive        time.Duration     //
	Retry            int               // 重试次数
	Backoff          Backoff           // 每次timeout后等待时间,nil不等待
	ContentType      string            // 编码格式
	Charset          string            // 编码格式,utf-8,GBK
	Header           http.Header       // 消息头
	Query            url.Values        // 查询参数
	Cookies          []*http.Cookie    //
	Datas            map[string]string // 用户扩展字段
	Hooks            Hooks             //
}

func (o *Options) setNewDefault() {
	o.DialTimeout = defaultDialTimeout
	o.KeepAlive = defaultKeepAlive
	o.HandshakeTimeout = defaultHandshakeTimeout
}

func (o *Options) build(opts ...Option) {
	o.Timeout = defaultTimeout
	o.ContentType = defaultContentType
	o.Context = context.Background()
	o.Backoff = defaultBackoff
	for _, fn := range opts {
		fn(o)
	}
}

func (o *Options) toRawQuery(query url.Values) string {
	for k, v := range o.Query {
		for _, x := range v {
			query.Add(k, x)
		}
	}

	return query.Encode()
}

func (o *Options) AddHeader(key string, value interface{}) {
	if o.Header == nil {
		o.Header = make(http.Header)
	}

	addValue(o.Header, key, value)
}

func (o *Options) AddHeaders(headers map[string]string) {
	if o.Header == nil {
		o.Header = make(http.Header)
	}
	addValueMap(o.Query, headers)
}

func (o *Options) AddQuery(key string, value interface{}) {
	if o.Query == nil {
		o.Query = make(url.Values)
	}
	addValue(o.Query, key, value)
}

func (o *Options) AddQueries(queries map[string]string) {
	if o.Query == nil {
		o.Query = make(url.Values)
	}
	addValueMap(o.Query, queries)
}

func (o *Options) AddCookie(cookie *http.Cookie) {
	o.Cookies = append(o.Cookies, cookie)
}

func (o *Options) AddCookies(cookies []*http.Cookie) {
	o.Cookies = append(o.Cookies, cookies...)
}

func (o *Options) AddData(k, v string) {
	if o.Datas == nil {
		o.Datas = make(map[string]string)
	}
	o.Datas[k] = v
}

func (o *Options) AddDatas(datas map[string]string) {
	if o.Datas == nil {
		o.Datas = make(map[string]string)
	}

	for k, v := range datas {
		o.Datas[k] = v
	}
}

func (o *Options) AddHook(hook Hook) {
	o.Hooks = append(o.Hooks, hook)
}

func (o *Options) AddHooks(hooks []Hook) {
	o.Hooks = append(o.Hooks, hooks...)
}

func (o *Options) AddAuthorization(auth string) {
	o.AddHeader("Authorization", auth)
}

func (o *Options) AddBasicAuth(username, password string) {
	o.AddHeader("Authorization", "Basic "+basicAuth(username, password))
}

func (o *Options) AddBearAuth(auth string) Option {
	return func(o *Options) {
		o.AddHeader("Authorization", "Bearer "+auth)
	}
}

func (o *Options) AddXJwtToken(token string) Option {
	return func(o *Options) {
		o.AddHeader("X-Jwt-Token", token)
	}
}

func (o *Options) AddXAuthToken(token string) Option {
	return func(o *Options) {
		o.AddHeader("X-Auth-Token", token)
	}
}

/////////////////////////////////////////////
// Option func
/////////////////////////////////////////////
func WithOptions(opts *Options) Option {
	return func(o *Options) {
		*o = *opts
	}
}

func WithContext(c context.Context) Option {
	return func(o *Options) {
		o.Context = c
	}
}

func WithBaseURL(baseURL string) Option {
	return func(o *Options) {
		o.BaseURL = baseURL
	}
}

func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

func WithDialTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = t
	}
}

func WithHandshakeTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.HandshakeTimeout = t
	}
}

func WithKeepAlive(t time.Duration) Option {
	return func(o *Options) {
		o.KeepAlive = t
	}
}

func WithRetry(r int) Option {
	return func(o *Options) {
		o.Retry = r
	}
}

func WithBackoff(b Backoff) Option {
	return func(o *Options) {
		o.Backoff = b
	}
}

func WithContentType(ct string) Option {
	return func(o *Options) {
		o.ContentType = ct
	}
}

func WithCharset(t string) Option {
	return func(o *Options) {
		o.Charset = t
	}
}

func WithHeader(key string, value interface{}) Option {
	return func(o *Options) {
		o.AddHeader(key, value)
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.AddHeaders(headers)
	}
}

func WithQuery(key string, value interface{}) Option {
	return func(o *Options) {
		o.AddQuery(key, value)
	}
}

func WithQueries(queries map[string]string) Option {
	return func(o *Options) {
		o.AddQueries(queries)
	}
}

func WithCookie(cookie *http.Cookie) Option {
	return func(o *Options) {
		o.AddCookie(cookie)
	}
}

func WithCookies(cookies []*http.Cookie) Option {
	return func(o *Options) {
		o.AddCookies(cookies)
	}
}

func WithAuthorization(auth string) Option {
	return func(o *Options) {
		o.AddAuthorization(auth)
	}
}

func WithBasicAuth(username, password string) Option {
	return func(o *Options) {
		o.AddBasicAuth(username, password)
	}
}

func WithBearAuth(auth string) Option {
	return func(o *Options) {
		o.AddBearAuth(auth)
	}
}

func WithXJwtToken(token string) Option {
	return func(o *Options) {
		o.AddXJwtToken(token)
	}
}

func WithXAuthToken(token string) Option {
	return func(o *Options) {
		o.AddXAuthToken(token)
	}
}
