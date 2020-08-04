package ghttp

import (
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()
	var text string
	if _, err := c.Get("http://www.baidu.com", &text); err != nil {
		t.Log(err)
	} else {
		t.Log(text)
	}
}
