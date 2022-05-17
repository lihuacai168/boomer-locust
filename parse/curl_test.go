package parse

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type M map[string]interface{}

func TestParse(t *testing.T) {
	addSample(t, `curl 'http://121.5.2.74/auth/listUser' \
  -H 'Accept: */*' \
  -H 'Accept-Language: zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-HK;q=0.6' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -H 'Referer: http://121.5.2.74/' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36' \
  -H 'token: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE2NTI5NTM2MzMsImlkIjo5MSwidXNlcm5hbWUiOiJ0ZXN0ZXIiLCJuYW1lIjoiXHU2ZDRiXHU4YmQ1XHU4ZDI2XHU1M2Y3MTExIiwiZW1haWwiOiIxMjM0NUBxcS5jb20iLCJyb2xlIjoxLCJwaG9uZSI6InF3ZTExMSIsImNyZWF0ZWRfYXQiOiIyMDIxLTExLTI1IDIzOjM2OjU0IiwidXBkYXRlZF9hdCI6IjIwMjItMDQtMjcgMjA6MTc6MzYiLCJkZWxldGVkX2F0IjowLCJ1cGRhdGVfdXNlciI6OTEsImxhc3RfbG9naW5fYXQiOiIyMDIyLTA1LTE3IDE3OjQ3OjEzIiwiYXZhdGFyIjoiaHR0cHM6Ly9zdGF0aWMucGl0eS5mdW4vYXZhdGFyL3VzZXJfOTEuanBnIiwiaXNfdmFsaWQiOnRydWV9.D63FHiF-pYGEpSUJrCtBHj-qvSdyDxSkAXodxPwIciw' \
  --compressed \
  --insecure`, M{
		"method": "GET",
		"url":    "http://121.5.2.74/auth/listUser",
	})
	addSample(t, "curl -XPUT http://api.sloths.com/sloth/4", M{
		"method": "PUT",
		"url":    "http://api.sloths.com/sloth/4",
	})

	addSample(t, "curl http://api.sloths.com", M{
		"method": "GET",
		"url":    "http://api.sloths.com",
	})

	addSample(t, "curl -H \"Accept-Encoding: gzip\" --compressed http://api.sloths.com", M{
		"method": "GET",
		"url":    "http://api.sloths.com",
		"header": M{
			"Accept-Encoding": "gzip",
		},
	})

	addSample(t, "curl -X DELETE http://api.sloths.com/sloth/4", M{
		"method": "DELETE",
		"url":    "http://api.sloths.com/sloth/4",
	})

	addSample(t, "curl -d \"foo=bar\" https://api.sloths.com", M{
		"method": "POST",
		"url":    "https://api.sloths.com",
		"header": M{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		"body": "foo=bar",
	})

	addSample(t, "curl -H \"Accept: text/plain\" --header \"User-Agent: slothy\" https://api.sloths.com", M{
		"method": "GET",
		"url":    "https://api.sloths.com",
		"header": M{
			"Accept":     "text/plain",
			"User-Agent": "slothy",
		},
	})

	addSample(t, "curl --cookie 'species=sloth;type=galactic' slothy https://api.sloths.com", M{
		"method": "GET",
		"url":    "https://api.sloths.com",
		"header": M{
			"Cookie": "species=sloth;type=galactic",
		},
	})

	addSample(t, "curl --location --request GET 'http://api.sloths.com/users?token=admin'", M{
		"method": "GET",
		"url":    "http://api.sloths.com/users?token=admin",
	})
}

func addSample(t *testing.T, url string, exp M) {
	request, _ := Parse(url)
	check(t, exp, request)
}

func check(t *testing.T, exp M, got *Request) {
	for key, value := range exp {
		switch key {
		case "method":
			assert.Equal(t, value, got.Method)
		case "url":
			assert.Equal(t, value, got.Url)
		case "body":
			assert.Equal(t, value, got.Body)
		case "header":
			headers := value.(M)
			for k, v := range headers {
				assert.Equal(t, v, got.Header[k])
			}
		}
	}
}
