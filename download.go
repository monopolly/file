package file

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/monopolly/useragent"
	"github.com/valyala/fasthttp"
)

func httpclient(proxy ...string) *http.Client {
	transport := &http.Transport{
		Dial:                (&net.Dialer{Timeout: 10 * time.Second}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	if len(proxy) > 0 {
		p, err := url.ParseRequestURI(proxy[0])
		if err == nil {
			transport.Proxy = http.ProxyURL(p)
		}
	}

	return &http.Client{Timeout: time.Second * 10, Transport: transport}
}

// добавляет хедеры и генерит юзер агента как реальный юзер
// proxy: http://proxyIp:proxyPort
func Get(link string, proxy ...string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	//req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-us")
	req.Header.Add("DNT", "1")
	req.Header.Add("User-agent", useragent.Generate())
	c := httpclient(proxy...)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code is %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func Download(link string) (body []byte) {
	body, _ = Get(link)
	return
}

func Post(link string, body []byte) (resp *fasthttp.Response) {
	req := fasthttp.AcquireRequest()
	resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(link)
	req.Header.SetMethod(http.MethodPost)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-us")
	req.Header.Add("DNT", "1")
	req.Header.Add("User-agent", useragent.Generate())
	req.SetBody(body)
	_ = fasthttp.Do(req, resp)
	return
}

// возвращает финальный урл если есть редиректы
func Redirect(link string) (real string, err error) {
	req, err := http.NewRequest(http.MethodHead, link, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-us")
	req.Header.Add("DNT", "1")
	req.Header.Add("User-agent", useragent.Generate())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	real = resp.Request.URL.String()
	return
}
