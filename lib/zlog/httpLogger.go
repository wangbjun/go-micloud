package zlog

import (
	"net/http"
	"time"
)

// 在默认http client基础上增加日志功能
var HttpLoggerTransport = &loggedRoundTripper{http.DefaultTransport}

type loggedRoundTripper struct {
	rt http.RoundTripper
}

func (c *loggedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	Logger.Sugar().Infof("request_start method = %s url = %s", request.Method, request.URL.String())

	startTime := time.Now()

	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36")

	response, err := c.rt.RoundTrip(request)

	duration := time.Since(startTime)
	duration /= time.Millisecond

	if err != nil {
		Logger.Sugar().Errorf("response_error method = %s duration = %d url = %s error = %s",
			request.Method, duration, request.URL.String(), err.Error())
	} else {
		Logger.Sugar().Infof("response_success method = %s status = %d duration = %d url = %s",
			request.Method, response.StatusCode, duration, request.URL.String())
	}

	return response, err
}
