package js_engine

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"nasu/config"
	"net"
	"net/http"
	"syscall"
	"time"
)

// HTTP constructor to return a fresh HTTP client for the dev
func (j *JS_ENGINE) HTTP(m, u string) *HTTP_CLIENT {
	return &HTTP_CLIENT{
		js_engine: j,
		method:    m,
		url:       u,
		headers:   make(map[string]string),
		body:      "",
	}
}

// HTTP_CLIENT is a wrapper for making GET and POST requests
type HTTP_CLIENT struct {
	js_engine   *JS_ENGINE
	method      string
	url         string
	headers     map[string]string
	body        string
	running     bool
	resp        *HTTP_RESPONSE
	waitTimeout bool
}

type HTTP_RESPONSE struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	Header  http.Header
	Trailer http.Header

	Body          string
	ContentLength int64

	TimedOut bool
}

// Set is a stub for testing
func (h *HTTP_CLIENT) Request() *HTTP_CLIENT {
	/*
		NOTE: I could write code here to throttle requests.
	*/

	h.running = true
	go func() {
		// We setup a custom dialer. The special sauce is in httpSocketControl()
		safeDialer := &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
			Control:   httpSocketControl,
		}

		safeTransport := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           safeDialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          1,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		// We finally get the HTTP client here and set the transport for security
		client := http.Client{
			Timeout:   time.Duration(5 * time.Second),
			Transport: safeTransport,
		}

		req, err := http.NewRequest(h.method, h.url, bytes.NewBufferString(h.body))
		if err != nil {
			h.js_engine.Errors = append(h.js_engine.Errors, err.Error())
			h.js_engine.VM.Interrupt("halt")
			return
		}

		for v, n := range h.headers {
			req.Header.Set(n, v)
		}
		req.Header.Set("X-Abuse-Info", config.CONFIG.AbuseHeader)

		resp, err := client.Do(req)
		if err != nil {
			h.js_engine.Errors = append(h.js_engine.Errors, err.Error())
			h.js_engine.VM.Interrupt("halt")
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			h.js_engine.Errors = append(h.js_engine.Errors, err.Error())
			h.js_engine.VM.Interrupt("halt")
			return
		}

		h.resp = &HTTP_RESPONSE{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Proto:      resp.Proto,
			ProtoMajor: resp.ProtoMajor,
			ProtoMinor: resp.ProtoMinor,

			Header:  resp.Header,
			Trailer: resp.Trailer,

			Body:          string(body),
			ContentLength: resp.ContentLength,

			TimedOut: false,
		}
		h.running = false
	}()
	return h
}

func (h *HTTP_CLIENT) Wait(timeout_ms int64) *HTTP_RESPONSE {
	if timeout_ms == 0 {
		timeout_ms = config.CONFIG.MaxWaitMs
	}
	if config.CONFIG.MaxWaitMs != 0 && timeout_ms > config.CONFIG.MaxWaitMs {
		timeout_ms = config.CONFIG.MaxWaitMs
	}
	h.waitTimeout = false
	start_time := time.Now()
	duration := time.Since(start_time)

	for h.running && !h.waitTimeout {
		duration = time.Since(start_time)
		if duration.Milliseconds() >= timeout_ms {
			h.waitTimeout = true
			h.resp.TimedOut = true
		}
	}
	return h.resp
}

func (h *HTTP_CLIENT) SetHeader(n, v string) *HTTP_CLIENT {
	h.headers[n] = v
	return h
}

func (h *HTTP_CLIENT) SetBody(b string) *HTTP_CLIENT {
	h.body = b
	return h
}

func isAllowedIPAddress(address net.IP) bool {
	for _, not_allowed_ip := range config.CONFIG.NetBlacklist {
		if not_allowed_ip.Contains(address) {
			return false
		}
	}
	return true
}

func httpSocketControl(network string, address string, conn syscall.RawConn) error {
	if !(network == "tcp4" || network == "tcp6") {
		return fmt.Errorf("%s is not a safe network type", network)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%s is not a valid host/port pair: %s", address, err)
	}

	ipaddress := net.ParseIP(host)
	if ipaddress == nil {
		return fmt.Errorf("%s is not a valid IP address", host)
	}

	if !isAllowedIPAddress(ipaddress) {
		return fmt.Errorf("connecting out to %s is forbidden.", ipaddress)
	}

	if !(port == "80" || port == "443") {
		return fmt.Errorf("%s is not a safe port number", port)
	}

	return nil
}
