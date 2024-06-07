/*
 * @Author: flwfdd
 * @Date: 2024-06-06 17:18:15
 * @LastEditTime: 2024-06-07 11:46:14
 * @Description: 封装的网络请求工具包
 * _(:з」∠)_
 */
package request

import (
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	net_url "net/url"
	"strings"
)

type Response struct {
	Code    int
	Text    string
	Content []byte
	Header  http.Header
}

func request(request_type string, url string, headers map[string]string, request_body io.Reader) (Response, error) {
	req, _ := http.NewRequest(request_type, url, request_body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				for _, cookie := range via[len(via)-1].Cookies() {
					req.AddCookie(cookie)
				}
			}
			return nil
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{res.StatusCode, string(body), body, res.Header}, nil
}

func Get(url string, headers map[string]string) (Response, error) {
	return request("GET", url, headers, nil)
}

func Post(url string, headers map[string]string) (Response, error) {
	return request("POST", url, headers, nil)
}

func PostForm(url string, form map[string]string, headers map[string]string) (Response, error) {
	formValues := net_url.Values{}
	for k, v := range form {
		formValues.Set(k, v)
	}
	formDataStr := formValues.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return request("POST", url, headers, formBytesReader)
}

func PostJSON(url string, data string, headers map[string]string) (Response, error) {
	headers["Content-Type"] = "application/json; charset=UTF-8"
	return request("POST", url, headers, strings.NewReader(data))
}

func Put(url string, data []byte, headers map[string]string) (Response, error) {
	return request("PUT", url, headers, bytes.NewBuffer(data))
}
