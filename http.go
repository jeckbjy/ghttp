package ghttp

import "net/http"

var Default = NewClient()

func Post(url string, req interface{}, result interface{}, opts ...Option) (*http.Response, error) {
	return Default.Post(url, req, result, opts...)
}

func Get(url string, result interface{}, opts ...Option) (*http.Response, error) {
	return Default.Get(url, result, opts...)
}
