package httpclient

import (
	"errors"
	"net"
	"net/http"
	"time"
)

type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type Client struct {
    inner *http.Client
    retryDelays []time.Duration
}

var DefaultRetryDelays = []time.Duration {
    200 * time.Millisecond,
    500 * time.Millisecond,
    1 *time.Second,
}

func NewDefaultClient() *Client {
    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
        KeepAlive: 30 * time.Second,
    }

    transport := &http.Transport{
        DialContext: dialer.DialContext,
        ForceAttemptHTTP2: true,
        MaxIdleConns: 100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout: 90 * time.Second,
        TLSHandshakeTimeout: 5 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    }

    httpClient := &http.Client{
        Transport: transport,
        Timeout: 15 * time.Second,
    }

    return &Client {
        inner: httpClient,
        retryDelays: DefaultRetryDelays,
    }
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
    return c.inner.Do(req)
}

func isRetryableStatus(statusCode int) bool {
    return statusCode == http.StatusBadGateway || statusCode == http.StatusServiceUnavailable ||
         statusCode == http.StatusGatewayTimeout
}

func isRetryableError(err error) bool {
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Timeout() {
        return true
    }

    return false
}

func (c *Client) DoWithRetry(req *http.Request) (*http.Response, error) {
    var lastErr error

    for attempt, delay :=  range append([]time.Duration{0}, c.retryDelays...) {
        if attempt > 0 {
            time.Sleep(delay)
        }

        resp, err := c.inner.Do(req)
        if err == nil {
            if !isRetryableStatus(resp.StatusCode) {
                return resp, nil
            }

            _ = resp.Body.Close()
            lastErr = errors.New(resp.Status)
            continue
        }

        if isRetryableError(err) {
            lastErr = err
            continue
        }

        return nil, err
    }

    if lastErr == nil {
        lastErr = errors.New("request failed after retries")
    }
    return nil, lastErr
}