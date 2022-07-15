package file

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/monopolly/useragent"
	"github.com/valyala/fasthttp"
)

func Download(link string) (file []byte, err error) {
	transport := &http.Transport{Dial: (&net.Dialer{Timeout: 20 * time.Second}).Dial, TLSHandshakeTimeout: 20 * time.Second}
	c := &http.Client{Timeout: time.Second * 20, Transport: transport}
	resp, err := c.Get(link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	file, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

//добавляет хедеры и генерит юзер агента как реальный юзер
func Downloads(link string) (b []byte, err error) {

	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Languag", "en-us")
	req.Header.Add("DNT", "1")
	/* генерим агента */
	req.Header.Add("User-agent", useragent.Generate())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Status code is %d", resp.StatusCode)
		return
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func Post(link string, body []byte) (resp *fasthttp.Response) {
	req := fasthttp.AcquireRequest()
	resp = fasthttp.AcquireResponse()
	req.SetRequestURI(link)
	req.Header.SetMethod("POST")
	req.SetBody(body)
	fasthttp.Do(req, resp)
	return
}

//возвращает финальный урл если есть редиректы
func Redirect(link string) (real string, err error) {

	req, _ := http.NewRequest("HEAD", link, nil)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Languag", "en-us")
	req.Header.Add("DNT", "1")
	/* генерим агента */
	req.Header.Add("User-agent", useragent.Generate())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	real = resp.Request.URL.String()
	return
}
