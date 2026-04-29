package file

import (
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/monopolly/useragent"
)

/*  */
func DownloadFile(from string, to string, headers map[string]string, progress ...func(now, total, percent int)) (err error) {

	ua := useragent.Generate()
	p := func(now, total, percent int) {}
	req, err := grab.NewRequest(to, from)
	if err != nil {
		return
	}

	if headers != nil {
		for k, v := range headers {
			req.HTTPRequest.Header.Add(k, v)
		}

		if userAgent := headers["User-agent"]; userAgent != "" {
			ua = userAgent
		}
	}

	req.HTTPRequest.Header.Add("Connection", "keep-alive")
	req.HTTPRequest.Header.Add("Accept", "*/*")
	req.HTTPRequest.Header.Add("Accept-Language", "en-us")
	req.HTTPRequest.Header.Add("DNT", "1")

	c := grab.NewClient()
	c.UserAgent = ua
	resp := c.Do(req)

	if len(progress) > 0 {
		p = progress[0]
	}

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			p(int(resp.BytesComplete()), int(resp.Size()), int(100*resp.Progress()))
		case <-resp.Done:
			break Loop
		}
	}

	return resp.Err()
}
